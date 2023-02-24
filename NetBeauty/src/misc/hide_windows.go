package misc

import (
	"syscall"
)

// @reference https://github.com/exercism/cli/blob/052030145d92b0777a808b1348b91478cabd77c0/visibility/hide_file_windows.go

func IsHiddenFile(file string) (bool, error) {
	ptr, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return false, err
	}

	isHidden := (attributes & syscall.FILE_ATTRIBUTE_HIDDEN) == syscall.FILE_ATTRIBUTE_HIDDEN

	return isHidden, nil
}

func HideFile(file string) error {
	return setVisibility(file, false)
}

func ShowFile(file string) error {
	return setVisibility(file, true)
}

func setVisibility(file string, visible bool) error {
	ptr, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return err
	}

	attributes, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return err
	}

	if visible {
		attributes &^= syscall.FILE_ATTRIBUTE_HIDDEN
	} else {
		attributes |= syscall.FILE_ATTRIBUTE_HIDDEN
	}

	return syscall.SetFileAttributes(ptr, attributes)
}
