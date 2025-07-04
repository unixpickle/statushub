//go:build linux
// +build linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

const ioctlReadTermios = unix.TCGETS
const ioctlWriteTermios = unix.TCSETS

func disableEcho(f *os.File) error {
	fd := int(f.Fd())

	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return err
	}

	termios.Lflag &^= unix.ECHO
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return err
	}

	return nil
}
