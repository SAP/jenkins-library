<<<<<<< HEAD
// +build !windows

package cmd

import "syscall"

func init() {
	// unset umask otherwise permissions on file or directory
	// creation are altered in unpredictable ways.
	syscall.Umask(0)
}
||||||| parent of 2c68cedf (Add genrated files)
=======
//go:build !windows
// +build !windows

package cmd

import "syscall"

func init() {
	// unset umask otherwise permissions on file or directory
	// creation are altered in unpredictable ways.
	syscall.Umask(0)
}
>>>>>>> 2c68cedf (Add genrated files)
