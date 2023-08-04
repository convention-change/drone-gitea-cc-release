package gitea_cc_release_plugin

import (
	"fmt"
	"github.com/sinlov/drone-info-tools/drone_info"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsBuildDebugOpen(c *cli.Context) bool {
	return c.Bool("config.debug") || c.Bool(drone_info.NameCliStepsDebug)
}

// BindCliFlag
// check args here
func BindCliFlag(c *cli.Context, cliVersion, cliName string, drone drone_info.Drone) (*Plugin, error) {
	debug := IsBuildDebugOpen(c)

	rootFolderPath := c.String(NameRootFolderPath)
	if rootFolderPath == "" {
		wdPath, errGetWd := os.Getwd()
		if errGetWd != nil {
			return nil, errGetWd
		}
		rootFolderPath = wdPath
	}

	changeLogFullPath := filepath.Join(rootFolderPath, c.String(NameReadChangeLogFile))

	baseUrl := c.String(NameApiBaseUrl)

	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl + "/"
	}

	note := c.String(NameNote)
	if note != "" {
		noteContent, err := readStringOrFile(note)
		if err != nil {
			return nil, err
		}
		note = noteContent
	}

	config := Config{
		Debug:         debug,
		TimeoutSecond: c.Uint("config.timeout_second"),

		RootFolderPath: rootFolderPath,

		GiteaBaseUrl:  baseUrl,
		GiteaInsecure: c.Bool(NameGiteaInsecure),
		GiteaApiKey:   c.String(NameGiteaApiKey),

		GiteaReleaseFileGlobs: c.StringSlice(NameReleaseFiles),
		FilesChecksum:         c.StringSlice(NameFilesChecksum),
		GiteaFileExistsDo:     c.String(NameFileExistsDo),
		GiteaDraft:            c.Bool(NameDraft),
		GiteaPrerelease:       c.Bool(NamePrerelease),
		GiteaTitle:            c.String(NameTitle),
		GiteaNote:             note,

		NoteByConventionChange: c.Bool(NameNoteByConventionChange),
		ReadChangeLogFile:      changeLogFullPath,

		GitRemote: c.String(NameGitRemote),
	}

	if config.Debug {
		drone_log.ShowLogLineNo(true)
		for _, e := range os.Environ() {
			log.Println(e)
		}
	}

	drone_log.Debugf("args config.timeout_second: %v", config.TimeoutSecond)

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
			Name:       NameApiBaseUrl,
			Usage:      "gitea base url",
			HasBeenSet: false,
			EnvVars:    []string{EnvApiBaseUrl},
		},
		&cli.BoolFlag{
			Name:    NameGiteaInsecure,
			Usage:   "visit base-url via insecure https protocol",
			EnvVars: []string{EnvGiteaInsecure},
		},
		&cli.StringFlag{
			Name:       NameGiteaApiKey,
			Usage:      "gitea api key",
			HasBeenSet: false,
			EnvVars:    []string{EnvGiteaApiKey},
		},

		&cli.StringSliceFlag{
			Name:    NameReleaseFiles,
			Usage:   "release as files by glob pattern",
			EnvVars: []string{EnvReleaseFiles},
		},
		&cli.StringSliceFlag{
			Name:    NameFilesChecksum,
			Usage:   fmt.Sprintf("generate specific checksums support: %v", CheckSumSupport),
			EnvVars: []string{EnvFilesChecksum},
		},
		&cli.StringFlag{
			Name:    NameFileExistsDo,
			Usage:   fmt.Sprintf("what to do if file already exist support: %v", supportFileExistsDoList),
			Value:   FileExistsDoFail,
			EnvVars: []string{EnvFileExistsDo},
		},
		&cli.StringFlag{
			Name:    NameTitle,
			Usage:   "release title if not set will use tag name",
			EnvVars: []string{EnvTitle},
		},
		&cli.StringFlag{
			Name:    NameNote,
			Usage:   "release note will try read file, if set open note by convention change will cover this",
			Value:   "",
			EnvVars: []string{EnvNote},
		},

		&cli.BoolFlag{
			Name:    NameNoteByConventionChange,
			Usage:   "release note by convention change",
			Value:   false,
			EnvVars: []string{EnvNoteByConventionChange},
		},
		&cli.StringFlag{
			Name:    NameReadChangeLogFile,
			Usage:   "read change log file",
			Value:   "CHANGELOG.md",
			EnvVars: []string{EnvReadChangeLogFile},
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
		&cli.StringFlag{
			Name:    NameRootFolderPath,
			Usage:   "root folder path if not set will get cwd folder",
			Hidden:  true,
			EnvVars: []string{EnvRootFolderPath},
		},
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

		&cli.BoolFlag{
			Name:    NameDraft,
			Usage:   "draft release",
			Value:   true,
			EnvVars: []string{EnvDraft},
		},
		&cli.BoolFlag{
			Name:    NamePrerelease,
			Usage:   "set the release as prerelease",
			Value:   true,
			EnvVars: []string{EnvPrerelease},
		},
	}
}

func readStringOrFile(input string) (string, error) {
	// Check if input is a file path
	if _, err := os.Stat(input); err != nil && os.IsNotExist(err) {
		// No file found => use input as result
		return input, nil
	} else if err != nil {
		return "", err
	}
	result, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
