package browser

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "syscall"
)

// Open attempts to open the url in a browser cross-platform.
// If incognito is true, it tries to launch a private/incognito window.
// When running as root with sudo, it will drop privileges to the invoking
// user (SUDO_USER) to launch the browser in their desktop session.
func Open(url string, incognito bool) error {
    var cmd *exec.Cmd

    switch runtime.GOOS {
    case "linux":
        // Prefer explicit browsers that support incognito/private flags
        if bin := firstOnPath("google-chrome-stable", "google-chrome", "chromium", "chromium-browser", "brave-browser", "microsoft-edge", "microsoft-edge-stable"); bin != "" {
            args := []string{"--new-window"}
            if incognito {
                args = append(args, "--incognito")
            }
            args = append(args, url)
            cmd = exec.Command(bin, args...)
        } else if bin := firstOnPath("firefox"); bin != "" {
            args := []string{}
            if incognito {
                args = append(args, "-private-window")
            }
            args = append(args, url)
            cmd = exec.Command(bin, args...)
        } else {
            // Fallback: xdg-open (no incognito support)
            cmd = exec.Command("xdg-open", url)
        }
        dropToSudoUser(cmd)

    case "darwin":
        // macOS: use `open -na` to target a specific app with args
        if incognito {
            // Try Chrome first
            if appExists("/Applications/Google Chrome.app") || appExists(filepath.Join(os.Getenv("HOME"), "Applications/Google Chrome.app")) {
                cmd = exec.Command("open", "-na", "Google Chrome", "--args", "--incognito", url)
            } else if appExists("/Applications/Firefox.app") || appExists(filepath.Join(os.Getenv("HOME"), "Applications/Firefox.app")) {
                cmd = exec.Command("open", "-na", "Firefox", "--args", "-private-window", url)
            } else {
                // Fallback without incognito
                cmd = exec.Command("open", url)
            }
        } else {
            cmd = exec.Command("open", url)
        }
        dropToSudoUser(cmd)

    case "windows":
        // Try Chrome/Edge/Firefox with private flags; else fallback to shell handler
        if bin := firstOnPath("chrome", "chrome.exe", "msedge", "msedge.exe"); bin != "" && incognito {
            cmd = exec.Command(bin, "--incognito", url)
        } else if bin := firstOnPath("firefox", "firefox.exe"); bin != "" && incognito {
            cmd = exec.Command(bin, "-private-window", url)
        } else {
            cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
        }
    default:
        return errors.New("unsupported platform for auto-open")
    }
    return cmd.Start()
}

func firstOnPath(candidates ...string) string {
    for _, c := range candidates {
        if p, err := exec.LookPath(c); err == nil && p != "" {
            return p
        }
    }
    return ""
}

func appExists(path string) bool {
    if st, err := os.Stat(path); err == nil && st.IsDir() {
        return true
    }
    return false
}

// dropToSudoUser modifies cmd to run as the SUDO_USER if the current process
// is running with uid 0. It also adjusts common environment variables so the
// desktop session (DBus/XDG) can be discovered.
func dropToSudoUser(cmd *exec.Cmd) {
    if runtime.GOOS == "windows" {
        return
    }
    if os.Geteuid() != 0 {
        return
    }
    sudoUser := os.Getenv("SUDO_USER")
    if sudoUser == "" {
        return
    }
    u, err := user.Lookup(sudoUser)
    if err != nil {
        return
    }
    uid, _ := strconv.Atoi(u.Uid)
    gid, _ := strconv.Atoi(u.Gid)
    cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}

    env := os.Environ()
    env = setEnv(env, "HOME", u.HomeDir)
    env = setEnv(env, "USER", sudoUser)
    env = setEnv(env, "LOGNAME", sudoUser)
    if runtime.GOOS == "linux" {
        xdg := fmt.Sprintf("/run/user/%d", uid)
        if st, err := os.Stat(xdg); err == nil && st.IsDir() {
            env = setEnv(env, "XDG_RUNTIME_DIR", xdg)
            if !hasEnv(env, "DBUS_SESSION_BUS_ADDRESS") {
                env = append(env, "DBUS_SESSION_BUS_ADDRESS=unix:path="+filepath.Join(xdg, "bus"))
            }
        }
    }
    cmd.Env = env
}

func setEnv(env []string, key, value string) []string {
    prefix := key + "="
    for i, e := range env {
        if strings.HasPrefix(e, prefix) {
            env[i] = prefix + value
            return env
        }
    }
    return append(env, prefix+value)
}

func hasEnv(env []string, key string) bool {
    prefix := key + "="
    for _, e := range env {
        if strings.HasPrefix(e, prefix) {
            return true
        }
    }
    return false
}

