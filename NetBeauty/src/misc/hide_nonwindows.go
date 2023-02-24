// +build !windows

package misc

func IsHiddenFile(file string) (bool, error) {
	return false, nil
}

func HideFile(file string) error {
	return setVisibility(file, false)
}

func ShowFile(file string) error {
	return setVisibility(file, true)
}

func setVisibility(file string, visible bool) error {
	return nil
}
