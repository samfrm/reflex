package browser

import "os/exec"

// On Windows, do nothing special; the browser will open under the current user.
func dropToSudoUser(cmd *exec.Cmd) { /* no-op */ }
