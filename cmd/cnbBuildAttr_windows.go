//go:build windows

package cmd

import (
	"syscall"
)

func getSysProcAttr(_ int, _ int) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
