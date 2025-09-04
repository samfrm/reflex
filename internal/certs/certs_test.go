package certs

import (
    "os"
    "testing"
)

func TestIsMkcertInstalledRespectsPATH(t *testing.T) {
    old := os.Getenv("PATH")
    defer os.Setenv("PATH", old)
    // Empty PATH should cause lookup to fail and return false.
    os.Setenv("PATH", "")
    if IsMkcertInstalled() {
        t.Fatalf("expected IsMkcertInstalled to be false with empty PATH")
    }
}

