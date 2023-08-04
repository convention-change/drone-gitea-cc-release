package gitea_cc_release_plugin

import (
	"code.gitea.io/sdk/gitea"
	"crypto/tls"
	"fmt"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"path/filepath"
)

var (
	ErrMissingTag = fmt.Errorf("newReleaseClient missing tag")
)

// Release holds ties the drone env data and gitea client together.
type releaseClient struct {
	client *gitea.Client
	owner  string
	repo   string
	tag    string
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

func (r *releaseClient) buildRelease() (*gitea.Release, error) {
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

func (r *releaseClient) uploadFiles(releaseID int64) error {
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

func newReleaseClient(drone drone_info.Drone, config Config) (pluginReleaseClient, error) {

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
		client:       client,
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

type pluginReleaseClient interface {
	Title() string

	SetTitle(title string)

	Tag() string

	SetNote(noteContent string)

	buildRelease() (*gitea.Release, error)

	uploadFiles(releaseID int64) error
}
