package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
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

func ReadAllDir(dir string) (paths []string, err error) {
	fd, err := ioutil.ReadDir(dir)
	paths = make([]string, 0)
	if err != nil {
		return paths, err
	}
	for _, fi := range fd {
		if fi.IsDir() {
			paths = append(paths, fi.Name())
		}
	}
	return paths, nil
}

func ReadAllFile(dir string) (paths []string, err error) {
	fd, err := ioutil.ReadDir(dir)
	paths = make([]string, 0)
	if err != nil {
		return paths, err
	}
	for _, fi := range fd {
		if !fi.IsDir() {
			paths = append(paths, fi.Name())
		}
	}
	return paths, nil
}

func GetAllFiles(dir string, recursive bool) []string {
	dir = filepath.Clean(dir)
	rd, _ := ioutil.ReadDir(dir)
	files := make([]string, 0)
	for _, fi := range rd {
		absName := dir + "/" + fi.Name()
		if fi.IsDir() {
			if recursive {
				files = append(files, GetAllFiles(absName, recursive)...)
			}
		} else {
			files = append(files, absName)
		}
	}
	return files
}

func GetFileMD5(file string) (string, error) {
	hash := md5.New()

	handle, error := os.Open(file)

	defer handle.Close()

	if error != nil {
		return "", error
	}

	_, error = io.Copy(hash, handle)

	if error != nil {
		return "", error
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GetStringMD5(str string) (string, error) {
	bytes := []byte(str)
	hash := md5.New()

	_, error := hash.Write(bytes)

	if error != nil {
		return "", error
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
