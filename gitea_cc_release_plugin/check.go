package gitea_cc_release_plugin

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"hash/adler32"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	CheckSumMd5     = "md5"
	CheckSumSha1    = "sha1"
	CheckSumSha256  = "sha256"
	CheckSumSha512  = "sha512"
	CheckSumAdler32 = "adler32"
	CheckSumCrc32   = "crc32"
	CheckSumBlake2b = "blake2b"
	CheckSumBlake2s = "blake2s"
)

var (
	CheckSumSupport = []string{
		CheckSumMd5,
		CheckSumSha1,
		CheckSumSha256,
		CheckSumSha512,
		CheckSumAdler32,
		CheckSumCrc32,
		CheckSumBlake2b,
		CheckSumBlake2s,
	}
)

func Checksum(r io.Reader, method string) (string, error) {
	b, err := io.ReadAll(r)

	if err != nil {
		return "", err
	}

	switch method {
	case CheckSumMd5:
		return fmt.Sprintf("%x", md5.Sum(b)), nil
	case CheckSumSha1:
		return fmt.Sprintf("%x", sha1.Sum(b)), nil
	case CheckSumSha256:
		return fmt.Sprintf("%x", sha256.Sum256(b)), nil
	case CheckSumSha512:
		return fmt.Sprintf("%x", sha512.Sum512(b)), nil
	case CheckSumAdler32:
		return strconv.FormatUint(uint64(adler32.Checksum(b)), 10), nil
	case CheckSumCrc32:
		return strconv.FormatUint(uint64(crc32.ChecksumIEEE(b)), 10), nil
	case CheckSumBlake2b:
		return fmt.Sprintf("%x", blake2b.Sum256(b)), nil
	case CheckSumBlake2s:
		return fmt.Sprintf("%x", blake2s.Sum256(b)), nil
	}

	return "", fmt.Errorf("hashing method %s is not supported", method)
}

func WriteChecksumsByFiles(files, methods []string, root string) ([]string, error) {
	if len(methods) == 0 {
		return nil, fmt.Errorf("no hashing methods specified")
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files specified")
	}
	checksums := make(map[string][]string)

	for _, method := range methods {
		for _, file := range files {
			handle, err := os.Open(file)

			if err != nil {
				return nil, fmt.Errorf("failed to read %s artifact: %s", file, err)
			}

			hash, err := Checksum(handle, method)

			if err != nil {
				return nil, err
			}

			checksums[method] = append(checksums[method], hash, file)
		}
	}

	if root != "" {
		if !strings.HasSuffix(root, string(filepath.Separator)) {
			root = fmt.Sprintf("%s%s", root, string(filepath.Separator))
		}
	}

	for method, results := range checksums {
		filename := method + "sum.txt"
		savePath := filepath.Join(root, filename)
		saveDir := filepath.Dir(savePath)
		if !pathExistsFast(saveDir) {
			errMkSaveDir := mkdir(saveDir)
			if errMkSaveDir != nil {
				return nil, errMkSaveDir
			}
		}
		f, err := os.Create(savePath)

		if err != nil {
			return nil, err
		}

		for i := 0; i < len(results); i += 2 {
			hash := results[i]
			file := results[i+1]
			if root != "" {
				file = strings.Replace(file, root, "", -1)
			}

			fileBaseName := filepath.Base(file)
			if _, errWrite := f.WriteString(fmt.Sprintf("%s  %s\n", hash, fileBaseName)); errWrite != nil {
				return nil, errWrite
			}
		}

		files = append(files, savePath)
	}

	return files, nil
}

var ErrGlobsEmpty = fmt.Errorf("globs is empty")

func FindFileByGlobs(globs []string, root string) ([]string, error) {
	if len(globs) == 0 {
		return nil, ErrGlobsEmpty
	}
	var findFiles []string
	if len(globs) > 0 {
		for _, glob := range globs {
			globed, errGlob := WalkAllByGlob(root, glob, true)
			if errGlob != nil {
				errGlobFind := fmt.Errorf("from glob find %s path: %s ,failed: %v", glob, root, errGlob)
				return nil, errGlobFind
			}
			if globed != nil {
				findFiles = append(findFiles, globed...)
			}
		}
	}

	return findFiles, nil
}

// WalkAllByGlob
// can walk all path then return as list, by glob with filepath.Glob
func WalkAllByGlob(path string, glob string, ignoreFolder bool) ([]string, error) {
	fiRoot, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("want Walk not exist at path: %s", path)
		}
		return nil, fmt.Errorf("want Walk not read at path: %s , err: %v", path, err)
	}
	if !fiRoot.IsDir() {
		return nil, fmt.Errorf("want Walk path is file, at: %s", path)
	}
	pathOfGlob := fmt.Sprintf("%s%s%s", path, `/`, glob)
	matches, err := filepath.Glob(pathOfGlob)
	if err != nil {
		return nil, fmt.Errorf("want Walk by path %s by glob %s ,err: %v", path, glob, err)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	files := make([]string, 0, 30)
	for _, match := range matches {
		f, errStat := os.Stat(match)
		if errStat != nil {
			return nil, fmt.Errorf("want Walk Stat at path %s by glob %s ,err: %v", match, glob, errStat)
		}
		if ignoreFolder && f.IsDir() {
			continue
		}
		files = append(files, match)
	}

	return files, nil
}

// pathExists path exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// pathExistsFast path exists fast
func pathExistsFast(path string) bool {
	exists, _ := pathExists(path)
	return exists
}

// mkdir will use FileMode 0766
func mkdir(path string) error {
	err := os.MkdirAll(path, os.FileMode(0766))
	if err != nil {
		return fmt.Errorf("fail MkdirAll at path: %s , err: %v", path, err)
	}
	return nil
}
