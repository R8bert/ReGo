package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// RunCommand executes a command and returns the result
func RunCommand(name string, args ...string) CommandResult {
	return RunCommandWithTimeout(name, 30*time.Second, args...)
}

// RunCommandWithTimeout executes a command with a timeout
func RunCommandWithTimeout(name string, timeout time.Duration, args ...string) CommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CommandResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// CommandExists checks if a command is available in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RunCommandLines executes a command and returns stdout split into lines
func RunCommandLines(name string, args ...string) ([]string, error) {
	result := RunCommand(name, args...)
	if result.Error != nil {
		return nil, fmt.Errorf("command failed: %w - %s", result.Error, result.Stderr)
	}

	if result.Stdout == "" {
		return []string{}, nil
	}

	return strings.Split(result.Stdout, "\n"), nil
}

// RunCommandWithEnv executes a command with additional environment variables
func RunCommandWithEnv(env map[string]string, name string, args ...string) CommandResult {
	cmd := exec.Command(name, args...)

	// Build environment
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CommandResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
