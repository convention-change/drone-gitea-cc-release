package gitea_cc_release_plugin

import (
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const (
	EnvApiKey  = "PLUGIN_RELEASE_GITEA_API_KEY"
	NameApiKey = "config.release_gitea_api_key"

	EnvFileKey  = "PLUGIN_RELEASE_GITEA_FILES"
	NameFileKey = "config.release_gitea_files"

	EnvGitRemote  = "PLUGIN_RELEASE_GIT_REMOTE"
	NameGitRemote = "config.release_git_remote"
)

// BindCliFlag
// check args here
func BindCliFlag(c *cli.Context, cliVersion, cliName string, drone drone_info.Drone) (*Plugin, error) {
	config := Config{
		Debug:         c.Bool("config.debug"),
		TimeoutSecond: c.Uint("config.timeout_second"),

		GiteaApiKey:       c.String(NameApiKey),
		GiteaReleaseFiles: c.StringSlice(NameFileKey),

		GitRemote: c.String(NameGitRemote),
	}

	drone_log.Debugf("args config.timeout_second: %v", config.TimeoutSecond)

	if config.Debug {
		for _, e := range os.Environ() {
			log.Println(e)
		}
	}

	//if config.Webhook == "" {
	//	err := fmt.Errorf("missing webhook, please set webhook env: %s", EnvWebHook)
	//	drone_log.Error(err)
	//	return nil, exit_cli.Err(err)
	//}

	// set default TimeoutSecond
	if config.TimeoutSecond == 0 {
		config.TimeoutSecond = 10
	}

	p := Plugin{
		Name:    cliName,
		Version: cliVersion,
		Drone:   drone,
		Config:  config,
	}
	return &p, nil
}

// Flag
// set flag at here
func Flag() []cli.Flag {
	return []cli.Flag{
		// plugin start
		&cli.StringFlag{
			Name:       NameApiKey,
			Usage:      "gitea api key",
			HasBeenSet: false,
			EnvVars:    []string{EnvApiKey},
		},
		&cli.StringSliceFlag{
			Name:    NameFileKey,
			Usage:   "release as files by glob pattern",
			EnvVars: []string{EnvFileKey},
		},
		&cli.StringFlag{
			Name:    NameGitRemote,
			Usage:   "release as files by glob pattern",
			EnvVars: []string{EnvGitRemote},
			Value:   "origin",
		},
		// plugin end
		//&cli.StringFlag{
		//	Name:    "config.new_arg,new_arg",
		//	Usage:   "",
		//	EnvVars: []string{"PLUGIN_new_arg"},
		//},
		// file_browser_plugin end
	}
}

// HideFlag
// set plugin hide flag at here
func HideFlag() []cli.Flag {
	return []cli.Flag{
		//&cli.UintFlag{
		//	Name:    "config.timeout_second,timeout_second",
		//	Usage:   "do request timeout setting second.",
		//	Hidden:  true,
		//	Value:   10,
		//	EnvVars: []string{"PLUGIN_TIMEOUT_SECOND"},
		//},
	}
}

// CommonFlag
// Other modules also have flags
func CommonFlag() []cli.Flag {
	return []cli.Flag{
		&cli.UintFlag{
			Name:    "config.timeout_second,timeout_second",
			Usage:   "do request timeout setting second.",
			Hidden:  true,
			Value:   10,
			EnvVars: []string{"PLUGIN_TIMEOUT_SECOND"},
		},
		&cli.BoolFlag{
			Name:    "config.debug,debug",
			Usage:   "debug mode",
			Value:   false,
			EnvVars: []string{drone_info.EnvKeyPluginDebug},
		},
	}
}
