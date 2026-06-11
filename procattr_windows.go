//go:build windows

package main

import "syscall"

// detachSysProcAttr starts the spawned daemon in a new process group so it is
// not killed when the parent console closes.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
