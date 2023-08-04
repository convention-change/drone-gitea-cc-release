package gitea_cc_release_plugin

import (
	"fmt"
	"github.com/convention-change/convention-change-log/changelog"
	"github.com/convention-change/convention-change-log/convention"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	droneStrTools "github.com/sinlov/drone-info-tools/tools/str_tools"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type (
	// Plugin plugin all config
	Plugin struct {
		Name    string
		Version string
		Drone   drone_info.Drone
		Config  Config
	}
)

func (p *Plugin) CleanResultEnv() error {
	for _, envItem := range cleanResultEnvList {
		err := os.Unsetenv(envItem)
		if err != nil {
			return fmt.Errorf("at FileBrowserPlugin.CleanResultEnv [ %s ], err: %v", envItem, err)
		}
	}
	return nil
}

func (p *Plugin) Exec() error {

	if p.Config.GiteaBaseUrl == "" {
		err := fmt.Errorf("missing git base url, please set env: %s", EnvApiBaseUrl)
		drone_log.Error(err)
		return exit_cli.Err(err)
	}

	if p.Config.GiteaApiKey == "" {
		err := fmt.Errorf("missing git api key, please set env: %s", EnvGiteaApiKey)
		drone_log.Error(err)
		return exit_cli.Err(err)
	}

	if !(droneStrTools.StrInArr(p.Config.GiteaFileExistsDo, supportFileExistsDoList)) {
		return exit_cli.Format("release_gitea_file_exists_do type only support %v", supportFileExistsDoList)
	}

	drone_log.Debugf("use GiteaBaseUrl: %v\n", p.Config.GiteaBaseUrl)
	drone_log.Debugf("use GiteaInsecure: %v\n", p.Config.GiteaInsecure)
	//drone_log.Debugf("use GiteaApiKey: %v\n", p.Config.GiteaApiKey)
	drone_log.Debugf("use GiteaReleaseFileGlobs: %v\n", p.Config.GiteaReleaseFileGlobs)
	drone_log.Debugf("use FilesChecksum: %v\n", p.Config.FilesChecksum)
	drone_log.Debugf("use NoteByConventionChange: %v\n", p.Config.NoteByConventionChange)

	rc, err := newReleaseClient(p.Drone, p.Config)
	if err != nil {
		drone_log.Error(err)
		return exit_cli.Err(err)
	}

	if p.Config.NoteByConventionChange {

		specFilePath := filepath.Join(p.Config.RootFolderPath, VersionRcFileName)
		changeLogSpec, errChangeLogSpecByPath := convention.LoadConventionalChangeLogSpecByPath(specFilePath)
		if errChangeLogSpecByPath != nil {
			drone_log.Error(errChangeLogSpecByPath)
			return exit_cli.Err(errChangeLogSpecByPath)
		}
		reader, errCC := changelog.NewReader(p.Config.ReadChangeLogFile, *changeLogSpec)
		if errCC == nil {
			rc.SetNote(reader.HistoryFirstContent())
			rc.SetTitle(reader.HistoryFirstTagShort())
		} else {
			drone_log.Warnf("not found change log or other error: %v\n", errCC)
		}
	}

	release, err := rc.buildRelease()
	if err != nil {
		drone_log.Error(err)
		return exit_cli.Err(err)
	}

	if errUpload := rc.uploadFiles(release.ID); errUpload != nil {
		drone_log.Error(errUpload)
		return exit_cli.Err(errUpload)
	}

	return nil
}

// randomStr
//
//	new random string by cnt
//
//nolint:golint,unused
func randomStr(cnt uint) string {
	var letters = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	result := make([]byte, cnt)
	keyL := len(letters)
	rand.Seed(time.Now().Unix())
	for i := range result {
		result[i] = letters[rand.Intn(keyL)]
	}
	return string(result)
}

// randomStrBySed
//
//	new random string by cnt and sed
//
//nolint:golint,unused
func randomStrBySed(cnt uint, sed string) string {
	var letters = []byte(sed)
	result := make([]byte, cnt)
	keyL := len(letters)
	rand.Seed(time.Now().Unix())
	for i := range result {
		result[i] = letters[rand.Intn(keyL)]
	}
	return string(result)
}

// setEnvFromStr
//
//	set env from string
//
//nolint:golint,unused
func setEnvFromStr(key string, val string) {
	err := os.Setenv(key, val)
	if err != nil {
		log.Fatalf("set env key [%v] string err: %v", key, err)
	}
}
