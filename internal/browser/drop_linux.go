package browser

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

// On Linux, drop privileges to the invoking sudo user so the browser opens
// in the user's desktop session. Also set XDG/DBus envs when possible.
func dropToSudoUser(cmd *exec.Cmd) {
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

	xdg := fmt.Sprintf("/run/user/%d", uid)
	if st, err := os.Stat(xdg); err == nil && st.IsDir() {
		env = setEnv(env, "XDG_RUNTIME_DIR", xdg)
		if !hasEnv(env, "DBUS_SESSION_BUS_ADDRESS") {
			env = append(env, "DBUS_SESSION_BUS_ADDRESS=unix:path="+filepath.Join(xdg, "bus"))
		}
	}
	cmd.Env = env
}
