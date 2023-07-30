package gitea_cc_release_plugin

import (
	"fmt"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sinlov-go/go-git-tools/git"
	"github.com/sinlov-go/go-git-tools/git_info"
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

	if p.Drone.Repo.SshUrl == "" {
		return exit_cli.Format("Drone.Repo.SshUrl is empty")
	}

	repositoryClone, err := git.NewRepositoryClone(memory.NewStorage(), nil,
		&goGit.CloneOptions{
			URL: p.Drone.Repo.SshUrl,
		},
	)
	if err != nil {
		drone_log.Warnf("clone repository SshUrl %s \nerr: %v", p.Drone.Repo.SshUrl, err)
	} else {
		commits, errLog := repositoryClone.Log("", "")
		if errLog != nil {
			drone_log.Warnf("get repositoryClone log err: %v", errLog)
		} else {
			drone_log.Infof("get repositoryClone commits len %d", len(commits))
		}
	}

	repositoryByPath, err := git.NewRepositoryByPath(p.Drone.Build.WorkSpace)
	if err != nil {
		return exit_cli.Format("at NewRepositoryByPath err: %v", err)
	}
	commits, errLog := repositoryByPath.Log("", "")
	if errLog != nil {
		drone_log.Warnf("get repositoryByPath log err: %v", errLog)
	} else {
		drone_log.Infof("get repositoryByPath commits len %d", len(commits))
	}
	tagLatestByCommitTime, errTagLatestByCommitTime := repositoryByPath.TagLatestByCommitTime()
	if errTagLatestByCommitTime != nil {
		drone_log.Warnf("get repositoryByPath TagLatestByCommitTime err: %v", errTagLatestByCommitTime)
	} else {
		drone_log.Infof("get repositoryByPath TagLatestByCommitTime.Name %v", tagLatestByCommitTime.Name)
	}
	commitLatestTagByTime, errCommitLatestTagByTime := repositoryByPath.CommitLatestTagByTime()
	if errCommitLatestTagByTime != nil {
		drone_log.Warnf("get repositoryByPath CommitLatestTagByTime err: %v", errCommitLatestTagByTime)
	} else {
		drone_log.Infof("get repositoryByPath CommitLatestTagByTime.Hash %v", commitLatestTagByTime.Hash.String())
	}

	fistRemoteInfo, err := git_info.RepositoryFistRemoteInfo(p.Drone.Build.WorkSpace, p.Config.GitRemote)
	if err != nil {
		drone_log.Warnf("at RepositoryFistRemoteInfo err: %v", err)
	} else {
		drone_log.Infof("fistRemoteInfo.Scheme %s", fistRemoteInfo.Scheme)
		drone_log.Infof("fistRemoteInfo.Host %s", fistRemoteInfo.Host)
		drone_log.Infof("fistRemoteInfo.User %s", fistRemoteInfo.User)
		drone_log.Infof("fistRemoteInfo.Repo %s", fistRemoteInfo.Repo)
		drone_log.Infof("fistRemoteInfo.Repo %s", fistRemoteInfo.UserInfo)
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
