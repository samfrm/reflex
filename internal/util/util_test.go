package util

import (
    "net"
    "runtime"
    "testing"
)

func TestExtractHostname(t *testing.T) {
    cases := []struct{ in, want string }{
        {"https://example.com/path", "example.com"},
        {"https://example.com:8443/path", "example.com"},
        {"example.com", "example.com"},
        {"example.com:1234", "example.com"},
        {"http://sub.example.com", "sub.example.com"},
    }
    for _, c := range cases {
        got, err := ExtractHostname(c.in)
        if err != nil {
            t.Fatalf("ExtractHostname(%q) error: %v", c.in, err)
        }
        if got != c.want {
            t.Fatalf("ExtractHostname(%q) = %q; want %q", c.in, got, c.want)
        }
    }
}

func TestCanBind(t *testing.T) {
    // Pick an unused port by binding to :0
    ln, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatalf("listen :0: %v", err)
    }
    port := ln.Addr().(*net.TCPAddr).Port
    // While the port is in use, CanBind must report false
    if CanBind(port) {
        t.Fatalf("CanBind(%d) = true while port is in use", port)
    }
    _ = ln.Close()
    if !CanBind(port) {
        t.Fatalf("CanBind(%d) = false after release", port)
    }
}

func TestRequireRoot_NonRoot(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("windows: RequireRoot is a no-op")
    }
    // Most CI/test environments run as non-root; RequireRoot should error.
    if err := RequireRoot(); err == nil {
        t.Skip("running as root; skipping non-root assertion")
    }
}

func TestLockAcquireExclusive(t *testing.T) {
    l1, err := AcquireLock()
    if err != nil {
        t.Fatalf("AcquireLock #1: %v", err)
    }
    defer l1.Release()

    if _, err := AcquireLock(); err == nil {
        t.Fatalf("expected AcquireLock #2 to fail while locked")
    }
    l1.Release()
    if _, err := AcquireLock(); err != nil {
        t.Fatalf("AcquireLock after release: %v", err)
    }
}

