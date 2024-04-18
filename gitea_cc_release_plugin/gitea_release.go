package gitea_cc_release_plugin

import (
	"bytes"
	"code.gitea.io/sdk/gitea"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sinlov-go/go-common-lib/pkg/filepath_plus"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	"golang.org/x/mod/module"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrMissingTag              = fmt.Errorf("NewReleaseClientByDrone missing tag, please check drone now in tag build")
	ErrPackageNotExist         = fmt.Errorf("PackageFetch not exist, code 404")
	ErrPathCanNotLoadGoModFile = fmt.Errorf("path can not load go.mod")
	ErrPackageGoExists         = fmt.Errorf("package go exists")
)

// Release holds ties the drone env data and gitea client together.
type releaseClient struct {
	client *gitea.Client
	debug  bool
	dryRun bool

	url        string
	ctx        context.Context
	mutex      *sync.RWMutex
	httpClient *http.Client

	accessToken string // this not in RWLock
	username    string
	password    string
	otp         string
	sudo        string

	owner string
	repo  string
	tag   string
	// tagTarget
	//is the branch or commit sha to tag
	tagTarget  string
	draft      bool
	prerelease bool
	// what to do if file already exist can use: overwrite, skip
	fileExistsDo string
	title        string
	note         string

	uploadFilePaths []string
	uploadDesc      string
}

type PackageGoInfo struct {
	ModVersion module.Version
}

// PackageGoUpload
// upload go package to gitea after gitea version 1.20.1
func (r *releaseClient) PackageGoUpload(rootPath string, removePath []string) (error, *PackageGoInfo) {

	errGoPkgFetch, exitsPackageInfo := r.PackageGoFetch(rootPath)
	if errGoPkgFetch != nil {
		if errGoPkgFetch == ErrPathCanNotLoadGoModFile {
			return fmt.Errorf("PackageGoUpload PackageGoFetch error: %s", errGoPkgFetch), nil
		}
		if errGoPkgFetch != ErrPackageNotExist {
			return fmt.Errorf("PackageGoUpload PackageGoFetch error: %s", errGoPkgFetch), nil
		}
	}
	if errGoPkgFetch == nil {
		return ErrPackageGoExists, exitsPackageInfo
	}

	if len(removePath) > 0 {
		drone_log.Debugf("try PackageGoUpload removePath: %v", removePath)
		for _, removePathItem := range removePath {
			removeFullPath := strings.Replace(removePathItem, "/", string(filepath.Separator), -1)
			removeFullPath = filepath.Join(rootPath, removeFullPath)
			errRemovePath := filepath_plus.RmDir(removeFullPath, true)
			if errRemovePath != nil {
				return fmt.Errorf("PackageGoUpload removePath: %s, error: %s", removeFullPath, errRemovePath), nil
			}
			drone_log.Infof("PackageGoUpload removePath success: %s", removeFullPath)
		}
	}

	outZipPath, modFile, errZipGoModPkg := CreateGoModZipFromDir(rootPath, r.tag)
	if errZipGoModPkg != nil {
		return errZipGoModPkg, nil
	}
	res := &PackageGoInfo{
		ModVersion: module.Version{
			Path:    modFile.Module.Mod.Path,
			Version: r.tag,
		},
	}

	drone_log.Debugf("try PackageGoUpload outZipPath: %s", outZipPath)

	fileBodyIO, errOpen := os.Open(outZipPath)
	if errOpen != nil {
		return fmt.Errorf("open zip file %s , error: %s", outZipPath, errOpen), res
	}
	defer func(fileBodyIO *os.File) {
		errFileBodyIO := fileBodyIO.Close()
		if errOpen != nil {
			log.Fatalf("try ResourcesPostFile file IO Close err: %v", errFileBodyIO)
		}
	}(fileBodyIO)

	uploadPath := fmt.Sprintf("/api/packages/%s/go/upload", r.owner)
	if r.dryRun {
		drone_log.Infof("PackageGoUpload dryRun: %s\n", outZipPath)
		drone_log.Infof("PackageGoUpload uploadPath: %s%s\n", r.url, uploadPath)
		return nil, res
	}
	statusCode, errPutGoPackage := r.getApiStatusCode("PUT", uploadPath, nil, fileBodyIO)
	if errPutGoPackage != nil {
		return fmt.Errorf("do PackageGoUpload go package [ %s ] err: %v", uploadPath, errPutGoPackage), res
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("do put go package [ %s ] errcode: %v, zip_path: %s", uploadPath, statusCode, outZipPath), res
	}

	return nil, res
}

func (r *releaseClient) PackageGoFetch(rootPath string) (error, *PackageGoInfo) {
	modFile, err := FetchGoModFile(rootPath)
	if err != nil {
		return ErrPathCanNotLoadGoModFile, nil
	}
	mod := modFile.Module.Mod
	goPackageName := mod.Path
	goPackageVersion := r.tag
	res := &PackageGoInfo{
		ModVersion: module.Version{
			Path:    goPackageName,
			Version: goPackageVersion,
		},
	}
	packageInfo, err := r.PackageFetch("go", goPackageName, goPackageVersion)
	if err != nil {
		return err, res
	}

	if packageInfo.Type != "go" || packageInfo.Name != res.ModVersion.Path || packageInfo.Version != res.ModVersion.Version {
		infoJson, errToJsonStr := dataJsonStr(packageInfo, true)
		if errToJsonStr != nil {
			return errToJsonStr, res
		}
		return fmt.Errorf("server package info not match local package info, data:\n%s", infoJson), res
	}

	return nil, res
}

type GiteaPackageInfo struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (r *releaseClient) PackageFetch(pkgType, name, version string) (*GiteaPackageInfo, error) {
	pkgName := name
	err := escapeValidatePathSegments(&pkgName)
	if err != nil {
		return nil, fmt.Errorf("PackageFetch escapeValidatePathSegments error: %s", err)
	}
	pkgVersion := version
	err = escapeValidatePathSegments(&pkgVersion)
	if err != nil {
		return nil, fmt.Errorf("PackageFetch escapeValidatePathSegments error: %s", err)
	}

	apiPath := fmt.Sprintf("/api/v1/packages/%s/%s/%s/%s", r.owner, pkgType, pkgName, pkgVersion)

	drone_log.Debugf("try PackageFetch apiPath: %s", apiPath)

	var giteaPackage GiteaPackageInfo
	resp, err := r.getApiParsedResponse("GET", apiPath, nil, nil, &giteaPackage)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, ErrPackageNotExist
		}
		return nil, fmt.Errorf("check package type [ %s ] [ %s ] err: %v", pkgType, apiPath, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check package type [ %s ] [ %s ] errcode: %v", pkgType, apiPath, resp.StatusCode)
	}
	return &giteaPackage, nil
}

func (r *releaseClient) BuildRelease() (*gitea.Release, error) {
	release, err := r.getRelease()

	if err != nil && release == nil {
		drone_log.Debugf("not getRelease release but can try new release, err: %v", err)
	} else if release != nil {
		drone_log.Infof("found Release ID:%d Draft:%v Prerelease:%v url: %s", release.ID, release.IsDraft, release.IsPrerelease, release.HTMLURL)
		return release, nil
	}

	if r.dryRun {
		drone_log.Infof("try to create release %s/%s\n", r.owner, r.repo)
		drone_log.Infof("dry run, not creating release\n")
		return &gitea.Release{
			ID: -1,
		}, nil
	}

	// if no release was found by that tag, create a new one
	release, err = r.newRelease()

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve or create a release: %s", err)
	}

	drone_log.Infof("new Release ID:%d Draft:%v Prerelease:%v url: %s", release.ID, release.IsDraft, release.IsPrerelease, release.HTMLURL)
	return release, nil
}

func (r *releaseClient) UploadFiles(releaseID int64) error {
	if len(r.uploadFilePaths) == 0 {
		if r.uploadDesc != "" {
			drone_log.Infof("%s\n", r.uploadDesc)
			return nil
		}
		drone_log.Infof("check settings no upload files found!\n")
		return nil
	}

	if r.dryRun {
		drone_log.Infof("try to upload files to release %s/%s\n", r.owner, r.repo)
		for _, filePath := range r.uploadFilePaths {
			drone_log.Infof("-> try upload file: %s\n", filePath)
		}
		drone_log.Infof("dry run, not uploading files\n")
		return nil
	}

	attachments, _, err := r.client.ListReleaseAttachments(r.owner, r.repo, releaseID, gitea.ListReleaseAttachmentsOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch existing assets: %s", err)
	}

	var uploadFiles []string

files:

	for _, filePath := range r.uploadFilePaths {
		for _, attachment := range attachments {
			if attachment.Name == filepath.Base(filePath) {
				switch r.fileExistsDo {
				case FileExistsDoOverwrite:
					// do nothing now we will delete the old file and upload the new one
				case FileExistsDoFail:
					return fmt.Errorf("asset file %s already exists", filepath.Base(filePath))
				case FileExistsDoSkip:
					drone_log.Infof("skipping pre-existing %s artifact\n", attachment.Name)
					continue files
				default:
					return fmt.Errorf("internal error, unkown file_exist value %s", r.fileExistsDo)
				}
			}
		}

		uploadFiles = append(uploadFiles, filePath)
	}

	for _, file := range uploadFiles {
		handle, errOpen := os.Open(file)

		if errOpen != nil {
			return fmt.Errorf("failed to read %s artifact: %s", file, errOpen)
		}

		fileBaseName := filepath.Base(file)

		for _, attachment := range attachments {
			if attachment.Name == fileBaseName {
				if _, err := r.client.DeleteReleaseAttachment(r.owner, r.repo, releaseID, attachment.ID); err != nil {
					return fmt.Errorf("failed to delete file base name: %s artifact: %s", fileBaseName, err)
				}

				drone_log.Infof("successfully deleted old attachment.ID[ %v ] artifact %s\n", attachment.ID, attachment.Name)
			}
		}

		if _, _, err = r.client.CreateReleaseAttachment(r.owner, r.repo, releaseID, handle, fileBaseName); err != nil {
			return fmt.Errorf("failed to upload file base name: %s artifact: %s", fileBaseName, err)
		}

		drone_log.Infof("successfully uploaded artifact file name [ %s ] path: %s \n", fileBaseName, file)
	}

	return nil
}

func (r *releaseClient) SetNote(noteContent string) {
	r.note = noteContent
}

func (r *releaseClient) Tag() string {
	return r.tag
}

func (r *releaseClient) SetTitle(nowTitle string) {
	r.title = nowTitle
}

func (r *releaseClient) Title() string {
	return r.title
}

func (r *releaseClient) getRelease() (*gitea.Release, error) {
	releases, _, err := r.client.ListReleases(r.owner, r.repo, gitea.ListReleasesOptions{})
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		if release.TagName == r.tag {
			drone_log.Debugf("Successfully retrieved %s release\n", r.tag)
			return release, nil
		}
	}
	return nil, fmt.Errorf("release %s not found", r.tag)
}

func (r *releaseClient) newRelease() (*gitea.Release, error) {
	c := gitea.CreateReleaseOption{
		TagName:      r.tag,
		Target:       r.tagTarget,
		IsDraft:      r.draft,
		IsPrerelease: r.prerelease,
		Title:        r.title,
		Note:         r.note,
	}

	release, _, err := r.client.CreateRelease(r.owner, r.repo, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create release: %s", err)
	}

	return release, nil
}

func (r *releaseClient) GetUploadDesc() string {
	return r.uploadDesc
}

// SetOTP sets OTP for 2FA
func (r *releaseClient) SetOTP(otp string) {
	r.mutex.Lock()
	r.otp = otp
	r.client.SetOTP(otp)
	r.mutex.Unlock()
}

// SetSudo sets username to impersonate.
func (r *releaseClient) SetSudo(sudo string) {
	r.mutex.Lock()
	r.sudo = sudo
	r.client.SetSudo(sudo)
	r.mutex.Unlock()
}

// SetBasicAuth sets username and password
func (r *releaseClient) SetBasicAuth(username, password string) {
	r.mutex.Lock()
	r.username, r.password = username, password
	r.client.SetBasicAuth(username, password)
	r.mutex.Unlock()
}

func NewReleaseClientByDrone(drone drone_info.Drone, config Config) (PluginReleaseClient, error) {

	if drone.Build.Tag == "" && !config.DryRun {
		return nil, ErrMissingTag
	}

	uploadDesc := ""
	var uploadFiles []string
	if len(config.GiteaReleaseFileGlobs) > 0 {
		findFiles, errGlobs := FindFileByGlobs(config.GiteaReleaseFileGlobs, config.GiteaReleaseFileGlobRootPath)
		if errGlobs != nil {
			return nil, errGlobs
		}

		if len(findFiles) == 0 {
			return nil, fmt.Errorf("not found files by globs: %v , at path: %s", config.GiteaReleaseFileGlobs, config.GiteaReleaseFileGlobRootPath)
		}

		repetitionFiles := findUploadFileRepetitionByBaseName(findFiles)
		if len(repetitionFiles) > 0 {
			return nil, fmt.Errorf("found files repetition by base name, now not support upload, repetition path as\n%s", strings.Join(repetitionFiles, "\n"))
		}

		uploadFiles = findFiles

		if len(config.FilesChecksum) > 0 {
			filesCheckRes, errCheckSum := WriteChecksumsByFiles(uploadFiles, config.FilesChecksum, config.GiteaReleaseFileGlobRootPath)

			if errCheckSum != nil {
				errCheckSumWrite := fmt.Errorf("from config.files_checksum failed: %v", errCheckSum)
				return nil, exit_cli.Err(errCheckSumWrite)
			}
			uploadFiles = filesCheckRes
		}
	} else {
		uploadDesc = "PLUGIN_RELEASE_GITEA_FILES not setting, skip upload files"
	}

	httpClient := &http.Client{}
	if config.GiteaInsecure {
		cookieJar, _ := cookiejar.New(nil)
		httpClient = &http.Client{
			Jar: cookieJar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	client, err := gitea.NewClient(config.GiteaBaseUrl, gitea.SetToken(config.GiteaApiKey), gitea.SetHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create gitea client: %s", err)
	}

	// if the title was not provided via .drone.yml we use the tag instead
	if config.GiteaTitle == "" {
		config.GiteaTitle = drone.Build.Tag
	}

	return &releaseClient{
		client: client,
		debug:  config.Debug,
		dryRun: config.DryRun,

		url:         strings.TrimSuffix(config.GiteaBaseUrl, "/"),
		ctx:         context.Background(),
		mutex:       &sync.RWMutex{},
		httpClient:  httpClient,
		accessToken: config.GiteaApiKey,

		owner:        drone.Repo.OwnerName,
		repo:         drone.Repo.ShortName,
		tag:          drone.Build.Tag,
		tagTarget:    drone.Build.RepoBranch,
		draft:        config.GiteaDraft,
		prerelease:   config.GiteaPrerelease,
		fileExistsDo: config.GiteaFileExistsDo,
		title:        config.GiteaTitle,
		note:         config.GiteaNote,

		uploadFilePaths: uploadFiles,
		uploadDesc:      uploadDesc,
	}, nil
}

type PluginReleaseClient interface {
	GetUploadDesc() string

	SetOTP(otp string)

	SetSudo(sudo string)

	SetBasicAuth(username, password string)

	Title() string

	SetTitle(title string)

	Tag() string

	SetNote(noteContent string)

	BuildRelease() (*gitea.Release, error)

	UploadFiles(releaseID int64) error

	PackageFetch(pkgType, name, version string) (*GiteaPackageInfo, error)

	PackageGoFetch(rootPath string) (error, *PackageGoInfo)

	PackageGoUpload(rootPath string, removePath []string) (error, *PackageGoInfo)
}

// giteaResponse represents the gitea response
type giteaResponse struct {
	*http.Response
}

func (r *releaseClient) getApiParsedResponse(method, path string, header http.Header, body io.Reader, obj interface{}) (*giteaResponse, error) {
	data, resp, err := r.getApiResponse(method, path, header, body)
	if err != nil {
		if r.debug {
			fmt.Printf("getApiParsedResponse code %d err: %s\n", resp.StatusCode, err)
		}
		return resp, err
	}
	return resp, json.Unmarshal(data, obj)
}

func (r *releaseClient) getApiResponse(method, path string, header http.Header, body io.Reader) ([]byte, *giteaResponse, error) {
	resp, err := r.doApiRequest(method, path, header, body)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// check for errors
	data, err := statusCodeToErr(resp)
	if err != nil {
		return data, resp, err
	}
	// success (2XX), read body
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, nil
}

func (r *releaseClient) getApiStatusCode(method, path string, header http.Header, body io.Reader) (int, error) {
	resp, err := r.doApiRequest(method, path, header, body)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func (r *releaseClient) doApiRequest(method, path string, header http.Header, body io.Reader) (*giteaResponse, error) {
	if r.client == nil {
		return nil, fmt.Errorf("gitea client is nil")
	}
	r.mutex.Lock()
	debug := r.debug
	urlFullPath := r.url + path
	if debug {
		fmt.Printf("%s: %s\nHeader: %v\nBody: %s\n", method, urlFullPath, header, body)
	}
	req, err := http.NewRequestWithContext(r.ctx, method, urlFullPath, body)
	if err != nil {
		r.mutex.RUnlock()
		return nil, err
	}

	if len(r.accessToken) != 0 {
		req.Header.Set("Authorization", "token "+r.accessToken)
	}
	if len(r.otp) != 0 {
		req.Header.Set("X-GITEA-OTP", r.otp)
	}
	if len(r.username) != 0 {
		req.SetBasicAuth(r.username, r.password)
	}
	if len(r.sudo) != 0 {
		req.Header.Set("Sudo", r.sudo)
	}

	for k, v := range header {
		req.Header[k] = v
	}

	r.mutex.Unlock()
	httpClient := r.httpClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if debug {
		fmt.Printf("Response: %v\n\n", resp)
	}

	return &giteaResponse{resp}, nil
}

// Converts a response for a HTTP status code indicating an error condition
// (non-2XX) to a well-known error value and response body. For non-problematic
// (2XX) status codes nil will be returned. Note that on a non-2XX response, the
// response body stream will have been read and, hence, is closed on return.
func statusCodeToErr(resp *giteaResponse) (body []byte, err error) {
	// no error
	if resp.StatusCode/100 == 2 {
		return nil, nil
	}

	//
	// error: body will be read for details
	//
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("body read on HTTP error %d: %v", resp.StatusCode, err)
	}

	switch resp.StatusCode {
	case http.StatusForbidden:
		return data, errors.New("403 Forbidden")
	case http.StatusNotFound:
		return data, errors.New("404 Not Found")
	case http.StatusConflict:
		return data, errors.New("409 Conflict")
	case http.StatusUnprocessableEntity:
		return data, fmt.Errorf("422 Unprocessable Entity: %s", string(data))
	}

	urlPath := resp.Request.URL.Path
	method := resp.Request.Method
	header := resp.Request.Header
	errMap := make(map[string]interface{})
	if err = json.Unmarshal(data, &errMap); err != nil {
		// when the JSON can't be parsed, data was probably empty or a
		// plain string, so we try to return a helpful error anyway
		return data, fmt.Errorf("Unknown API Error: %d\nRequest: '%s' with '%s' method '%s' header and '%s' body", resp.StatusCode, urlPath, method, header, string(data))
	}
	return data, errors.New(errMap["message"].(string))
}

// escapeValidatePathSegments is a help function to validate and encode url path segments
//
//nolint:golint,unused
func escapeValidateQuerySegments(seg ...*string) error {
	for i := range seg {
		if seg[i] == nil || len(*seg[i]) == 0 {
			return fmt.Errorf("path segment [%d] is empty", i)
		}
		*seg[i] = url.QueryEscape(*seg[i])
	}
	return nil
}

// escapeValidatePathSegments is a help function to validate and encode url path segments
//
//nolint:golint,unused
func escapeValidatePathSegments(seg ...*string) error {
	for i := range seg {
		if seg[i] == nil || len(*seg[i]) == 0 {
			return fmt.Errorf("path segment [%d] is empty", i)
		}
		*seg[i] = url.PathEscape(*seg[i])
	}
	return nil
}

func dataJsonStr(v any, beauty bool) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if beauty {
		var str bytes.Buffer
		errIndent := json.Indent(&str, data, "", "  ")
		if errIndent != nil {
			return "", errIndent
		}
		return str.String(), nil
	}
	return string(data), nil
}

// findUploadFileRepetitionByBaseName find upload file base name repetition
// return duplicates file full path, len 0 is not find
func findUploadFileRepetitionByBaseName(files []string) []string {
	seen := make(map[string]bool)
	var duplicates []string
	for _, fileFullPath := range files {
		fileBaseName := filepath.Base(fileFullPath)
		if seen[fileBaseName] {
			duplicates = append(duplicates, fileFullPath)
		} else {
			seen[fileBaseName] = true
		}
	}
	return duplicates
}
