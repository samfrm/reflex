package certs

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
)

// IsMkcertInstalled checks if mkcert is available on PATH.
func IsMkcertInstalled() bool {
    cmd := exec.Command("mkcert", "--version")
    if err := cmd.Run(); err != nil {
        return false
    }
    return true
}

// EnsureCertificates generates a domain certificate with mkcert in outDir.
// Returns paths to cert.pem and key.pem.
func EnsureCertificates(domain, outDir string) (string, string, error) {
    certPath := filepath.Join(outDir, "cert.pem")
    keyPath := filepath.Join(outDir, "key.pem")

    if fileExists(certPath) && fileExists(keyPath) {
        return certPath, keyPath, nil
    }

    cmd := exec.Command("mkcert", "-key-file", keyPath, "-cert-file", certPath, domain)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return "", "", fmt.Errorf("mkcert: %w", err)
    }
    return certPath, keyPath, nil
}

// EnsureCertificatesWithCAROOT is like EnsureCertificates but forces mkcert to
// use a specific CAROOT directory by setting the CAROOT environment variable.
func EnsureCertificatesWithCAROOT(domain, outDir, caroot string) (string, string, error) {
    certPath := filepath.Join(outDir, "cert.pem")
    keyPath := filepath.Join(outDir, "key.pem")

    if fileExists(certPath) && fileExists(keyPath) {
        return certPath, keyPath, nil
    }

    cmd := exec.Command("mkcert", "-key-file", keyPath, "-cert-file", certPath, domain)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    // Pin the CA root so root and user share the same CA
    if caroot != "" {
        cmd.Env = append(os.Environ(), "CAROOT="+caroot)
    }
    if err := cmd.Run(); err != nil {
        return "", "", fmt.Errorf("mkcert (CAROOT=%s): %w", caroot, err)
    }
    return certPath, keyPath, nil
}

// EnsureLocalCAInstalled runs `mkcert -install` to ensure the local CA is
// present in the system trust stores. It's safe to run multiple times.
func EnsureLocalCAInstalled() error {
    cmd := exec.Command("mkcert", "-install")
    out, err := cmd.CombinedOutput()
    s := string(out)
    // Detect a common non-fatal condition on fresh Linux installs where
    // browser security databases (NSS) aren't present yet.
    if strings.Contains(s, "no Firefox and/or Chrome/Chromium security databases found") {
        // Provide actionable hints while still treating installation as successful.
        // mkcert still installs the CA into the system trust store.
        fmt.Println("Note: mkcert did not find Firefox/Chrome security databases.")
        if runtime.GOOS == "linux" {
            fmt.Println("If you need trust in browsers using NSS, install NSS tools and re-run 'sudo mkcert -install'.")
            fmt.Println("Debian/Ubuntu: sudo apt-get install libnss3-tools")
            fmt.Println("Fedora: sudo dnf install nss-tools")
            fmt.Println("Arch: sudo pacman -S nss")
            fmt.Println("Also ensure you've launched Firefox/Chrome at least once to create a profile.")
        }
    }
    if err != nil {
        return fmt.Errorf("mkcert -install failed: %w\n%s", err, s)
    }
    return nil
}

func fileExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
}
