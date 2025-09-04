package hosts

import (
    "os"
    "path/filepath"
    "testing"
)

func TestAddRemoveContains(t *testing.T) {
    dir := t.TempDir()
    hp := filepath.Join(dir, "hosts")
    if err := os.WriteFile(hp, []byte("127.0.0.1 localhost\n"), 0o644); err != nil {
        t.Fatal(err)
    }
    m := Manager{Path: hp}

    if err := m.Add("127.0.0.1", "example.test"); err != nil {
        t.Fatalf("Add: %v", err)
    }
    present, line, err := m.Contains("example.test")
    if err != nil || !present {
        t.Fatalf("Contains: present=%v err=%v", present, err)
    }
    if line == "" {
        t.Fatalf("expected non-empty line")
    }
    if err := m.Remove("example.test"); err != nil {
        t.Fatalf("Remove: %v", err)
    }
    present, _, err = m.Contains("example.test")
    if err != nil {
        t.Fatalf("Contains: %v", err)
    }
    if present {
        t.Fatalf("expected not present")
    }
}

func TestAddDuplicateAndRemoveAllTagged(t *testing.T) {
    dir := t.TempDir()
    hp := filepath.Join(dir, "hosts")
    if err := os.WriteFile(hp, []byte("127.0.0.1 localhost\n"), 0o644); err != nil {
        t.Fatal(err)
    }
    m := Manager{Path: hp}

    if err := m.Add("127.0.0.1", "alpha.test"); err != nil {
        t.Fatalf("Add alpha: %v", err)
    }
    if err := m.Add("127.0.0.1", "alpha.test"); err == nil {
        t.Fatalf("expected duplicate add to fail with ErrAlreadyPresent")
    }
    if err := m.Add("127.0.0.1", "beta.test"); err != nil {
        t.Fatalf("Add beta: %v", err)
    }
    n, err := m.RemoveAllTagged()
    if err != nil {
        t.Fatalf("RemoveAllTagged: %v", err)
    }
    if n != 2 {
        t.Fatalf("RemoveAllTagged removed %d; want 2", n)
    }
}
