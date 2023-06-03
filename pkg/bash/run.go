package bash

import "os/exec"

// RunCommand executes a shell command.
func RunCommand(cmd string) ([]byte, error) {
	return exec.Command("sh", "-c", cmd).Output()
}
