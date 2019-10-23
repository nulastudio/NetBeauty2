package util

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func EnsureDirExists(dir string, perm os.FileMode) bool {
	if !PathExists(dir) {
		return os.MkdirAll(dir, perm) == nil
	} else {
		return os.Chmod(dir, perm) == nil
	}
}

func CopyFile(src string, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	dir := filepath.Dir(des)
	if !EnsureDirExists(dir, 0777) {
		return 0, errors.New("cannot create path: " + dir)
	}

	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}
