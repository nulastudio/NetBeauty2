package main

import (
	"syscall"
)

func Hide(file string) error {
	filenameW, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return err
	}
	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}
	return nil
}
