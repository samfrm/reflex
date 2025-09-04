package hosts

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "time"
)

const tag = "# reflex-managed"

var (
    ErrAlreadyPresent = errors.New("hosts entry already present")
)

// PathOrDefault returns the provided path or the OS-specific default.
func PathOrDefault(path string) string {
    if path != "" {
        return path
    }
    if runtime.GOOS == "windows" {
        return filepath.Clean(`C:\\Windows\\System32\\drivers\\etc\\hosts`)
    }
    return "/etc/hosts"
}

type Manager struct{ Path string }

// Add ensures a managed hosts entry exists for domain -> ip.
func (m Manager) Add(ip, domain string) error {
    if ip == "" || domain == "" {
        return fmt.Errorf("ip and domain required")
    }
    path := m.Path
    // Read current content
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    line := fmt.Sprintf("%s %s %s", ip, domain, tag)
    if strings.Contains(string(data), line) {
        return ErrAlreadyPresent
    }

    // Backup once per day with timestamp to be safe
    ts := time.Now().Format("20060102-150405")
    _ = os.WriteFile(path+".reflex."+ts+".bak", data, 0o644)

    f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    defer f.Close()
    if _, err := f.WriteString("\n" + line + "\n"); err != nil {
        return err
    }
    return nil
}

// Remove deletes the managed entry for the given domain (if present).
func (m Manager) Remove(domain string) error {
    if domain == "" {
        return fmt.Errorf("domain required")
    }
    path := m.Path
    in, err := os.Open(path)
    if err != nil {
        return err
    }
    defer in.Close()

    var b strings.Builder
    s := bufio.NewScanner(in)
    removed := false
    for s.Scan() {
        line := s.Text()
        if strings.Contains(line, tag) && strings.Contains(line, domain) {
            removed = true
            continue
        }
        b.WriteString(line)
        b.WriteByte('\n')
    }
    if err := s.Err(); err != nil {
        return err
    }
    if !removed {
        return nil
    }
    return os.WriteFile(path, []byte(strings.TrimRight(b.String(), "\n")+"\n"), 0o644)
}

// Contains reports if a managed entry exists for the domain.
func (m Manager) Contains(domain string) (bool, string, error) {
    data, err := os.ReadFile(m.Path)
    if err != nil {
        return false, "", err
    }
    lines := strings.Split(string(data), "\n")
    for _, l := range lines {
        if strings.Contains(l, tag) && strings.Contains(l, domain) {
            return true, l, nil
        }
    }
    return false, "", nil
}

// RemoveAllTagged removes all entries managed by Reflex, regardless of domain.
func (m Manager) RemoveAllTagged() (int, error) {
    data, err := os.ReadFile(m.Path)
    if err != nil {
        return 0, err
    }
    lines := strings.Split(string(data), "\n")
    var out []string
    removed := 0
    for _, l := range lines {
        if strings.Contains(l, tag) {
            removed++
            continue
        }
        out = append(out, l)
    }
    if removed == 0 {
        return 0, nil
    }
    s := strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
    return removed, os.WriteFile(m.Path, []byte(s), 0o644)
}
