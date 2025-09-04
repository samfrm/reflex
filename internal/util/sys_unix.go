//go:build !windows

package util

import "syscall"

func syscallKill(pid int, sig int) error {
    return syscall.Kill(pid, syscall.Signal(sig))
}

