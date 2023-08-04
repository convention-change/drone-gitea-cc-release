package cli

import (
	"fmt"
	"github.com/convention-change/drone-gitea-cc-release/gitea_cc_release_plugin"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2"
	"github.com/sinlov/drone-info-tools/drone_urfave_cli_v2/exit_cli"
	"github.com/sinlov/drone-info-tools/pkgJson"
	"github.com/urfave/cli/v2"
	"runtime"
	"time"
)

const (
	defaultExitCode    = 1
	CopyrightStartYear = "2023"
)

func NewCliApp() *cli.App {
	name := pkgJson.GetPackageJsonName()
	exit_cli.ChangeDefaultExitCode(defaultExitCode)
	if name == "" {
		panic("package.json name is empty")
	}
	app := cli.NewApp()
	app.Version = pkgJson.GetPackageJsonVersionGoStyle()
	app.Name = name
	app.Usage = pkgJson.GetPackageJsonDescription()
	year := time.Now().Year()
	jsonAuthor := pkgJson.GetPackageJsonAuthor()
	app.Copyright = fmt.Sprintf("© %s-%d %s by: %s, run on %s %s",
		CopyrightStartYear, year, jsonAuthor.Name, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	author := &cli.Author{
		Name:  jsonAuthor.Name,
		Email: jsonAuthor.Email,
	}
	app.Authors = []*cli.Author{
		author,
	}

	app.Before = GlobalBeforeAction
	app.Action = GlobalAction
	app.After = GlobalAfterAction

	flags := drone_urfave_cli_v2.UrfaveCliAppendCliFlag(drone_urfave_cli_v2.DroneInfoUrfaveCliFlag(), gitea_cc_release_plugin.CommonFlag())
	flags = drone_urfave_cli_v2.UrfaveCliAppendCliFlag(flags, gitea_cc_release_plugin.Flag())
	flags = drone_urfave_cli_v2.UrfaveCliAppendCliFlag(flags, gitea_cc_release_plugin.HideFlag())
	app.Flags = flags

	return app
}
