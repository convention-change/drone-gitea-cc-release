package gitea_cc_release_plugin_test

import (
	"fmt"
	"github.com/convention-change/drone-gitea-cc-release/gitea_cc_release_plugin"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"os"
	"testing"
)

func TestPlugin(t *testing.T) {
	// mock Plugin
	t.Logf("~> mock Plugin")
	p := gitea_cc_release_plugin.Plugin{
		Name:    mockName,
		Version: mockVersion,
	}
	// do Plugin
	t.Logf("~> do Plugin")
	if envCheck(t) {
		return
	}

	// use env:ENV_DEBUG
	p.Config.Debug = envDebug

	err := p.Exec()
	if nil == err {
		t.Fatalf("args [ %s ] empty error should be catch!", gitea_cc_release_plugin.EnvApiBaseUrl)
	}

	p.Config.GiteaBaseUrl = envApiBaseUrl

	err = p.Exec()
	if nil == err {
		t.Fatalf("args [ %s ] empty error should be catch!", gitea_cc_release_plugin.EnvGiteaApiKey)
	}

	p.Config.GiteaApiKey = envGiteaApiKey

	err = p.Exec()
	if nil == err {
		t.Fatalf("args [ %s ] error should be catch!", gitea_cc_release_plugin.EnvFileExistsDo)
	}

	p.Config.GiteaFileExistsDo = gitea_cc_release_plugin.FileExistsDoOverwrite

	err = p.Exec()
	if nil == err {
		t.Fatalf("args [ %s ] error should be catch!", drone_info.EnvDroneTag)
	}

	p.Config = gitea_cc_release_plugin.Config{
		Debug:         envDebug,
		TimeoutSecond: defTimeoutSecond,

		RootFolderPath: envProjectRoot,

		GiteaBaseUrl:  envApiBaseUrl,
		GiteaInsecure: envGiteaInsecure,
		GiteaApiKey:   envGiteaApiKey,

		GiteaReleaseFileGlobs: envReleaseFiles,
		FilesChecksum:         envFilesChecksum,

		GiteaFileExistsDo: envGiteaFileExistsDo,
		GiteaDraft:        envGiteaDraft,
		GiteaPrerelease:   envGiteaPrerelease,
		GiteaTitle:        envGiteaTitle,
		GiteaNote:         envGiteaNote,

		NoteByConventionChange: envNoteByConventionChange,
		ReadChangeLogFile:      envReadChangeLogFile,

		GitRemote: "origin",
	}

	droneMock, errMock := drone_info.MockDroneInfoByRefsAndNumber(
		envDroneProto,
		envDroneHost,
		envDroneHostName,
		envDroneProto,
		envDroneRemoteHost,
		envDroneRemoteOwner,
		envDroneRemoteRepoName,
		drone_info.DroneBuildStatusSuccess,
		"refs/tags/v1.0.0",
		1,
	)
	if errMock != nil {
		t.Fatal(errMock)
	}
	droneMock.Build.RepoBranch = "main"
	droneMock.Build.Tag = "v1.0.0"
	p.Drone = *droneMock

	globs, err := mockUploadFiles(t)
	if err != nil {
		t.Fatal(err)
	}

	p.Config.GiteaReleaseFileGlobs = globs
	p.Config.FilesChecksum = []string{
		gitea_cc_release_plugin.CheckSumMd5,
		gitea_cc_release_plugin.CheckSumSha256,
		gitea_cc_release_plugin.CheckSumSha512,
	}
	// verify Plugin
	err = p.Exec()
	if err != nil {
		t.Fatal(err)
	}

	err = p.CleanResultEnv()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", os.Getenv(gitea_cc_release_plugin.EnvPluginResultShareHost))
}

func mockUploadFiles(t *testing.T) ([]string, error) {

	var globsFiles = make(map[string][]string)
	globsFiles[fmt.Sprintf("*/%s/**.json.golden", t.Name())] = []string{
		"foo.json",
		"bar.json",
	}

	type testData struct {
		Name string
	}
	var fooData = testData{
		Name: "foo",
	}

	for _, values := range globsFiles {
		for _, value := range values {
			fooData.Name = value
			err := goldenDataSaveFast(t, fooData, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return maps.Keys(globsFiles), nil
}
