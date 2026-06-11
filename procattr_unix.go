//go:build !windows

package main

import "syscall"

// detachSysProcAttr detaches a spawned daemon into its own process group so it
// survives the parent exiting.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
