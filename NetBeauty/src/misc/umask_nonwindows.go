// +build !windows

package misc

import (
	"syscall"
)

func Umask() {
	syscall.Umask(0)
}
