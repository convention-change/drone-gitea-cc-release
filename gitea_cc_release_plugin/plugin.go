package gitea_cc_release_plugin

import (
	"fmt"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sinlov-go/go-git-tools/git"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	"log"
	"math/rand"
	"os"
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
	drone_log.Debugf("use GiteaApiKey: %v\n", p.Config.GiteaApiKey)
	drone_log.Debugf("use GiteaReleaseFiles: %v\n", p.Config.GiteaReleaseFiles)

	if p.Drone.Repo.HttpUrl == "" {
		return exit_cli.Format("Drone.Repo.HttpUrl is empty")
	}

	repositoryClone, err := git.NewRepositoryClone(memory.NewStorage(), nil,
		&goGit.CloneOptions{
			URL: p.Drone.Repo.HttpUrl,
		},
	)
	if err != nil {
		return exit_cli.Format("clone repository HttpUrl %s \nerr: %v", p.Drone.Repo.HttpUrl, err)
	}
	commits, err := repositoryClone.Log("", "")
	if err != nil {
		return exit_cli.Format("get repository log err: %v", err)
	}
	drone_log.Infof("get remote commits len %d", len(commits))
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
