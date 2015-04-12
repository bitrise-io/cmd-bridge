package main

import (
	// "errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	// "strings"
)

type CommandModel struct {
	Command          string `json:"command"`
	WorkingDirectory string `json:"working_directory"`
	LogFilePath      string `json:"log_file_path"`
}

func RunCommandInDirWithArgsAndWriters(dirPath string, command string, cmdArgs []string, stdOutWriter, stdErrWriter io.Writer) error {
	c := exec.Command(command, cmdArgs...)
	c.Stdout = stdOutWriter
	c.Stderr = stdErrWriter
	if dirPath != "" {
		c.Dir = dirPath
	}

	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func RunCommandInDirWithArgs(dirPath string, command string, cmdArgs []string) error {
	return RunCommandInDirWithArgsAndWriters(dirPath, command, cmdArgs, os.Stdout, os.Stderr)
}

func ExecuteUnlockKeychain(keychainName, keychainPsw string) error {
	cargs := []string{
		"-v", "unlock-keychain", "-p", keychainPsw, keychainName,
	}
	err := RunCommandInDirWithArgsAndWriters("", "security", cargs, CommandLogWriter, CommandLogWriter)
	return err
}

func ExecuteCommand(cmdToRun CommandModel) error {
	if err := WriteLineToCommandLog("[[command-start]]"); err != nil {
		return err
	}

	// // unlock keychain
	// if err := ExecuteUnlockKeychain(commandParams.KeychainName, commandParams.KeychainPassword); err != nil {
	// 	return err
	// }

	WriteLineToCommandLog(fmt.Sprintf("Command to run: $ %s", cmdToRun.Command))

	cargs := []string{
		"--login",
		"-c",
		cmdToRun.Command,
	}
	commandErr := RunCommandInDirWithArgsAndWriters(cmdToRun.WorkingDirectory, "bash", cargs, CommandLogWriter, CommandLogWriter)

	WriteLineToCommandLog("[[command-finished]]")
	return commandErr
}
