package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunCRM executes a crm-cli subcommand and returns stdout, stderr, and error.
func RunCRM(args ...string) (string, string, error) {
	cmd := exec.Command("crm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// RunCRMJSON executes crm with -f json and returns the raw JSON string.
func RunCRMJSON(args ...string) (string, error) {
	out, errStr, err := RunCRM(append(args, "-f", "json")...)
	if err != nil {
		if errStr != "" {
			return "", fmt.Errorf("%w: %s", err, errStr)
		}
		return "", err
	}
	return out, nil
}

// CRMExists checks if crm-cli is installed and reachable.
func CRMExists() bool {
	_, err := exec.LookPath("crm")
	return err == nil
}
