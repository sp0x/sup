package tools

import (
	"os/exec"
)

func HasBash() bool {
	path, err := exec.LookPath("bash")
	if err != nil {
		return false
	}
	return path != ""
}
