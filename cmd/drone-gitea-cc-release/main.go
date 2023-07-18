//go:build !test

package main

import (
	"github.com/convention-change/drone-gitea-cc-release"
	"github.com/convention-change/drone-gitea-cc-release/cmd/cli"
	"github.com/sinlov/drone-info-tools/drone_log"
	"github.com/sinlov/drone-info-tools/pkgJson"
	"github.com/sinlov/drone-info-tools/template"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	pkgJson.InitPkgJsonContent(drone_gitea_cc_release.PackageJson)
	template.RegisterSettings(template.DefaultFunctions)

	app := cli.NewCliApp()

	// kubernetes runner patch
	if _, err := os.Stat("/run/drone/env"); err == nil {
		errDotEnv := godotenv.Overload("/run/drone/env")
		if errDotEnv != nil {
			drone_log.Fatalf("load /run/drone/env err: %v", errDotEnv)
		}
	}

	// app run as urfave
	if err := app.Run(os.Args); nil != err {
		drone_log.Warnf("run err: %v", err)
	}
}
