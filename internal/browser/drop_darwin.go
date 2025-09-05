package browser

import (
	"os"
	"os/exec"
	"os/user"
)

// On macOS, adjust env to the sudo-invoking user so GUI apps attach
// to the right session. Avoid SysProcAttr credential fields (not portable).
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
	env := os.Environ()
	env = setEnv(env, "HOME", u.HomeDir)
	env = setEnv(env, "USER", sudoUser)
	env = setEnv(env, "LOGNAME", sudoUser)
	cmd.Env = env
}
