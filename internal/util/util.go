package util

import (
    "errors"
    "fmt"
    "net"
    neturl "net/url"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "time"
)

var (
    verbose bool
)

func EnableVerbose() { verbose = true }
func VLog(format string, args ...any) {
    if verbose {
        fmt.Printf(format+"\n", args...)
    }
}

// RequireRoot ensures the process is running with administrative privileges
// on Unix-like systems. On Windows it is a no-op.
func RequireRoot() error {
    if runtime.GOOS == "windows" {
        return nil
    }
    // On Unix, require effective uid 0 (sudo/root)
    if os.Geteuid() != 0 {
        return fmt.Errorf("reflex must be run with sudo/admin privileges to modify /etc/hosts and manage certificates.\nRe-run with sudo, e.g.: sudo %s", strings.Join(os.Args, " "))
    }
    return nil
}

// ExtractHostname returns the hostname part from a URL or raw host string.
func ExtractHostname(input string) (string, error) {
    if input == "" {
        return "", errors.New("empty input")
    }
    if u, err := neturl.Parse(input); err == nil && u.Host != "" {
        host := u.Host
        // Strip port if included
        h, _, err := net.SplitHostPort(host)
        if err == nil && h != "" {
            return h, nil
        }
        return host, nil
    }
    // Fallback: treat as hostname
    h, _, err := net.SplitHostPort(input)
    if err == nil && h != "" {
        return h, nil
    }
    return input, nil
}

// CanBind checks if a TCP port is available for binding on all interfaces.
func CanBind(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    _ = ln.Close()
    return true
}

// PathExists returns true if a path exists.
func PathExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
}

// Lock provides a simple cross-process lock using a file.
type Lock struct {
    path string
    f    *os.File
}

var lockOnce sync.Once

func lockPath() string {
    base := os.TempDir()
    if runtime.GOOS == "windows" {
        // Ensure a valid temp dir without colon in path for file name
        base = os.TempDir()
    }
    return filepath.Join(base, "reflex.lock")
}

// AcquireLock creates an exclusive lock file, failing if already locked.
func AcquireLock() (*Lock, error) {
    p := lockPath()
    // If a lock exists, check if it's stale
    if _, err := os.Stat(p); err == nil {
        if stale, _ := isLockStale(p); stale {
            _ = os.Remove(p)
        } else {
            return nil, fmt.Errorf("another reflex instance appears to be running (%s)", p)
        }
    }
    f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
    if err != nil {
        return nil, fmt.Errorf("another reflex instance appears to be running (%s)", p)
    }
    // Write PID and timestamp for diagnostics
    _, _ = f.WriteString(fmt.Sprintf("pid=%d\nstarted=%s\n", os.Getpid(), time.Now().Format(time.RFC3339)))
    return &Lock{path: p, f: f}, nil
}

func (l *Lock) Release() {
    if l == nil || l.f == nil {
        return
    }
    _ = l.f.Close()
    _ = os.Remove(l.path)
}

// RemoveLock deletes the lock file if present.
func RemoveLock() error { return os.Remove(lockPath()) }

// isLockStale attempts to decide if an existing lock belongs to a dead process.
// On Unix it reads a pid= line and probes with signal 0. On other OSes,
// fall back to age-based judgment (>24h old considered stale).
func isLockStale(path string) (bool, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return false, err
    }
    fi, _ := os.Stat(path)
    age := time.Duration(0)
    if fi != nil {
        age = time.Since(fi.ModTime())
    }
    if len(b) == 0 {
        // Consider very new empty files as in-progress; otherwise stale
        if age < 5*time.Second {
            return false, nil
        }
        return true, nil
    }
    var pid int
    for _, line := range strings.Split(string(b), "\n") {
        if strings.HasPrefix(line, "pid=") {
            p := strings.TrimPrefix(line, "pid=")
            if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
                pid = v
            }
            break
        }
    }
    if pid <= 0 {
        // Malformed or legacy lock without pid; treat as in-progress if very new, else stale
        if age < 5*time.Second {
            return false, nil
        }
        return true, nil
    }
    if pid > 0 && runtime.GOOS != "windows" {
        // On Unix, signal 0 checks for existence
        if err := unixSignal0(pid); err == nil {
            return false, nil // process exists
        }
        return true, nil // process missing
    }
    // Fallback: consider files older than 24h stale
    if fi, err := os.Stat(path); err == nil {
        if time.Since(fi.ModTime()) > 24*time.Hour {
            return true, nil
        }
    }
    return false, nil
}

// unixSignal0 sends signal 0 to pid on Unix; on non-Unix returns error.
func unixSignal0(pid int) error {
    if runtime.GOOS == "windows" {
        return fmt.Errorf("unsupported")
    }
    // Use syscall to send signal 0
    return syscallKill(pid, 0)
}
