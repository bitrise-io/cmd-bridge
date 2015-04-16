package main

import (
	// "errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	// "strings"
)

type EnvironmentKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CommandModel struct {
	Command          string                `json:"command"`
	WorkingDirectory string                `json:"working_directory"`
	LogFilePath      string                `json:"log_file_path"`
	Environments     []EnvironmentKeyValue `json:"environments"`
}

func RunCommandInDirWithArgsEnvsAndWriters(dirPath string, command string, cmdArgs []string, cmdEnvs []string, stdOutWriter, stdErrWriter io.Writer) error {
	c := exec.Command(command, cmdArgs...)
	c.Env = append(os.Environ(), cmdEnvs...)
	// c.Env = cmdEnvs // only the supported envs, no inherited ones
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

func ExecuteCommand(cmdToRun CommandModel) error {
	if err := WriteLineToCommandLog("[[command-start]]"); err != nil {
		return err
	}

	// // unlock keychain
	// if err := ExecuteUnlockKeychain(commandParams.KeychainName, commandParams.KeychainPassword); err != nil {
	// 	return err
	// }

	WriteLineToCommandLog(fmt.Sprintf("Command to run: $ %s", cmdToRun.Command))

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
	commandErr := RunCommandInDirWithArgsEnvsAndWriters(cmdToRun.WorkingDirectory, cmdExec, cmdArgs, cmdEnvs, CommandLogWriter, CommandLogWriter)

	if commandErr != nil {
		WriteLineToCommandLog(fmt.Sprintf("%s", commandErr))
	}

	WriteLineToCommandLog("[[command-finished]]")
	return commandErr
}
