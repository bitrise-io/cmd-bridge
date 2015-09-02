package main

import (
	// "errors"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	// "strings"
	"syscall"
)

// EnvironmentKeyValue ...
type EnvironmentKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CommandModel ...
type CommandModel struct {
	Command          string                `json:"command"`
	WorkingDirectory string                `json:"working_directory"`
	LogFilePath      string                `json:"log_file_path"`
	Environments     []EnvironmentKeyValue `json:"environments"`
}

// RunCommandInDirWithArgsEnvsAndWriters ...
func RunCommandInDirWithArgsEnvsAndWriters(dirPath string, command string, cmdArgs []string, cmdEnvs []string, stdOutWriter, stdErrWriter io.Writer) (int, error) {
	c := exec.Command(command, cmdArgs...)
	c.Env = append(os.Environ(), cmdEnvs...)
	// c.Env = cmdEnvs // only the supported envs, no inherited ones
	c.Stdout = stdOutWriter
	c.Stderr = stdErrWriter
	if dirPath != "" {
		c.Dir = dirPath
	}

	cmdExitCode := 0
	if err := c.Run(); err != nil {
		// Did the command fail because of an unsuccessful exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return 1, errors.New("Failed to cast exit status")
			}
			cmdExitCode = waitStatus.ExitStatus()
		}
		return cmdExitCode, err
	}
	return 0, nil
}

// func RunCommandInDirWithArgs(dirPath string, command string, cmdArgs []string) error {
// 	return RunCommandInDirWithArgsAndWriters(dirPath, command, cmdArgs, os.Stdout, os.Stderr)
// }

// func ExecuteUnlockKeychain(keychainName, keychainPsw string) error {
// 	cmdArgs := []string{
// 		"-v", "unlock-keychain", "-p", keychainPsw, keychainName,
// 	}
// 	err := RunCommandInDirWithArgsAndWriters("", "security", cmdArgs, CommandLogWriter, CommandLogWriter)
// 	return err
// }

// ExecuteCommand ...
func ExecuteCommand(cmdToRun CommandModel) (int, error) {
	if ConfigIsVerboseLogMode {
		if err := WriteLineToCommandLog("[[command-start]]"); err != nil {
			return 0, err
		}
	}

	// // unlock keychain
	// if err := ExecuteUnlockKeychain(commandParams.KeychainName, commandParams.KeychainPassword); err != nil {
	// 	return err
	// }

	if ConfigIsVerboseLogMode {
		if err := WriteLineToCommandLog(fmt.Sprintf("Command to run: $ %s", cmdToRun.Command)); err != nil {
			log.Println(" [!] Failed to write 'command to run' into Command Log")
		}
	}

	cmdExec := "/bin/bash"
	cmdArgs := []string{
		"--login",
		"-c",
		cmdToRun.Command,
	}
	cmdEnvs := []string{}
	envLength := len(cmdToRun.Environments)
	if envLength > 0 {
		cmdEnvs = make([]string, envLength, envLength)
		for idx, aEnvPair := range cmdToRun.Environments {
			cmdEnvs[idx] = aEnvPair.Key + "=" + aEnvPair.Value
		}
	}

	//
	cmdExitCode, commandErr := RunCommandInDirWithArgsEnvsAndWriters(cmdToRun.WorkingDirectory, cmdExec, cmdArgs, cmdEnvs, CommandLogWriter, CommandLogWriter)

	if commandErr != nil {
		if err := WriteLineToCommandLog(fmt.Sprintf("Command failed: %s", commandErr)); err != nil {
			log.Println(" [!] Failed to write 'Command failed' into Command Log")
		}
	}

	if ConfigIsVerboseLogMode {
		if err := WriteLineToCommandLog("[[command-finished]]"); err != nil {
			log.Println(" [!] Failed to write '[[command-finished]]' into Command Log")
		}
	}
	return cmdExitCode, commandErr
}
