//go:build !windows

package cmd

import "syscall"

func init() {
	// unset umask otherwise permissions on file or directory
	// creation are altered in unpredictable ways.
	syscall.Umask(0)
}
