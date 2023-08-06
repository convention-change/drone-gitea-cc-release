package gitea_cc_release_plugin

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/zip"
	"os"
	"path/filepath"
)

func FetchGoModFile(root string) (*modfile.File, error) {
	if !modfile.IsDirectoryPath(root) {
		return nil, fmt.Errorf("check error at CreateGoModZipFromDir not is go mod root path: %s", root)
	}
	goModPath := filepath.Join(root, "go.mod")

	goModData, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("check error at CreateGoModZipFromDir read go.mod: %s", err)
	}
	goModFile, err := modfile.Parse(goModPath, goModData, nil)
	if err != nil {
		return nil, fmt.Errorf("check error at CreateGoModZipFromDir parse go.mod: %s", err)
	}
	if goModFile == nil {
		return nil, fmt.Errorf("check error at CreateGoModZipFromDir parse go.mod is nil")
	}
	return goModFile, nil
}

// CreateGoModZipFromDir
// @doc https://go.dev/ref/mod#zip-files
//
//	root go mod root path
//	version go mod version
//
// return ( out zip path, modfile.File, error )
func CreateGoModZipFromDir(root string, version string) (string, *modfile.File, error) {

	goModFile, err := FetchGoModFile(root)
	if err != nil {
		return "", goModFile, fmt.Errorf("check error at CreateGoModZipFromDir check go.mod: %s", err)
	}
	goModule := goModFile.Module
	if goModule == nil {
		return "", goModFile, fmt.Errorf("check error at CreateGoModZipFromDir parse go.mod module is nil")
	}

	modVersion := goModule.Mod
	modVersion.Version = version

	outPath := filepath.Join(filepath.Dir(root), fmt.Sprintf("%s.zip", modVersion.Version))
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", goModFile, fmt.Errorf("check error at CreateGoModZipFromDir create zip file: %s", err)
	}

	err = zip.CreateFromDir(outFile, modVersion, root)
	if err != nil {
		return "", goModFile, fmt.Errorf("check error at CreateGoModZipFromDir CreateFromDir err: %s", err)
	}

	return outPath, goModFile, nil
}
