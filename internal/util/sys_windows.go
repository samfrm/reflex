//go:build windows

package util

import "fmt"

func syscallKill(pid int, sig int) error { return fmt.Errorf("unsupported") }

