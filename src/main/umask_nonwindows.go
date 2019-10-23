// +build !windows

package main

import (
	"syscall"
)

func Umask() {
	syscall.Umask(0)
}
