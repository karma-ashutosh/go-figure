package utils

import (
	"os/exec"
)

func ExecuteCommand(command string) string {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output) + "\nError: " + err.Error()
	}
	return string(output)
}
