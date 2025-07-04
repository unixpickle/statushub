//go:build !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !linux
// +build !darwin,!dragonfly,!freebsd,!netbsd,!openbsd,!linux

package main

import (
	"errors"
	"os"
)

func disableEcho(f *os.File) error {
	_ = f
	return errors.New("cannot disable echo")
}
