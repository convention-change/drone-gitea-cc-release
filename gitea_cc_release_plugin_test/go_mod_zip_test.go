package gitea_cc_release_plugin_test

import (
	"fmt"
	"github.com/convention-change/drone-gitea-cc-release"
	"github.com/convention-change/drone-gitea-cc-release/gitea_cc_release_plugin"
	"github.com/sinlov-go/go-common-lib/pkg/filepath_plus"
	"github.com/sinlov/drone-info-tools/pkgJson"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateGoModZipFromDir(t *testing.T) {
	t.Logf("~> mock CreateGoModZipFromDir")
	// mock CreateGoModZipFromDir

	pkgJson.InitPkgJsonContent(drone_gitea_cc_release.PackageJson)
	pkgVersion := pkgJson.GetPackageJsonVersionGoStyle()
	t.Logf("envProjectRoot: %s", envProjectRoot)
	t.Logf("pkgVersion: %s", pkgVersion)

	wantOutZip := filepath.Join(filepath.Dir(envProjectRoot), fmt.Sprintf("%s.zip", pkgVersion))
	if filepath_plus.PathExistsFast(wantOutZip) {
		err := os.Remove(wantOutZip)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Logf("~> do CreateGoModZipFromDir")
	// do CreateGoModZipFromDir
	outPath, err := gitea_cc_release_plugin.CreateGoModZipFromDir(envProjectRoot, pkgVersion)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("outPath: %s", outPath)

	// verify CreateGoModZipFromDir
	//assert.Equal(t, "expected", "")
}
