package gitea_cc_release_plugin

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var (
	ErrMissingTag = fmt.Errorf("NewReleaseClientByDrone missing tag")
)

// Release holds ties the drone env data and gitea client together.
type releaseClient struct {
	client     *gitea.Client
	debug      bool
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
}

func (r *releaseClient) PackageGoUpload(version string) error {
	//TODO implement me
	panic("implement me")
}

func (r *releaseClient) PackageGoFetch(version string) error {

	return nil
}

func (r *releaseClient) BuildRelease() (*gitea.Release, error) {
	release, err := r.getRelease()

	if err != nil && release == nil {
		drone_log.Debugf("not getRelease release but can try new release, err: %v", err)
	} else if release != nil {
		drone_log.Infof("found Release ID:%d Draft:%v Prerelease:%v url: %s", release.ID, release.IsDraft, release.IsPrerelease, release.HTMLURL)
		return release, nil
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
		drone_log.Infof("no upload files found\n")
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
					// do nothing
				case FileExistsDoFail:
					return fmt.Errorf("asset file %s already exists", path.Base(filePath))
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

		for _, attachment := range attachments {
			if attachment.Name == path.Base(file) {
				if _, err := r.client.DeleteReleaseAttachment(r.owner, r.repo, releaseID, attachment.ID); err != nil {
					return fmt.Errorf("failed to delete %s artifact: %s", file, err)
				}

				drone_log.Infof("successfully deleted old %s artifact\n", attachment.Name)
			}
		}

		if _, _, err = r.client.CreateReleaseAttachment(r.owner, r.repo, releaseID, handle, path.Base(file)); err != nil {
			return fmt.Errorf("failed to upload %s artifact: %s", file, err)
		}

		drone_log.Infof("successfully uploaded %s artifact\n", file)
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

	if drone.Build.Tag == "" {
		return nil, ErrMissingTag
	}

	var uploadFiles []string
	if len(config.GiteaReleaseFileGlobs) > 0 {
		findFiles, errGlobs := FindFileByGlobs(config.GiteaReleaseFileGlobs)
		if errGlobs != nil {
			return nil, errGlobs
		}
		uploadFiles = findFiles

		if len(config.FilesChecksum) > 0 {
			filesCheckRes, errCheckSum := WriteChecksumsByFiles(uploadFiles, config.FilesChecksum)

			if errCheckSum != nil {
				errCheckSumWrite := fmt.Errorf("from config.files_checksum failed: %v", errCheckSum)
				return nil, exit_cli.Err(errCheckSumWrite)
			}
			uploadFiles = filesCheckRes
		}
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
		client:      client,
		debug:       config.Debug,
		url:         config.GiteaBaseUrl,
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
	}, nil
}

type PluginReleaseClient interface {
	SetOTP(otp string)

	SetSudo(sudo string)

	SetBasicAuth(username, password string)

	Title() string

	SetTitle(title string)

	Tag() string

	SetNote(noteContent string)

	BuildRelease() (*gitea.Release, error)

	UploadFiles(releaseID int64) error

	PackageGoFetch(version string) error

	PackageGoUpload(version string) error
}

// giteaResponse represents the gitea response
type giteaResponse struct {
	*http.Response
}

func (r *releaseClient) getApiParsedResponse(method, path string, header http.Header, body io.Reader, obj interface{}) (*giteaResponse, error) {
	data, resp, err := r.getApiResponse(method, path, header, body)
	if err != nil {
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
	urlFullPath := r.url + "/api" + path
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
	case 403:
		return data, errors.New("403 Forbidden")
	case 404:
		return data, errors.New("404 Not Found")
	case 409:
		return data, errors.New("409 Conflict")
	case 422:
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
