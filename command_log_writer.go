package main

import (
	"fmt"
	"io"
	"os"
)

var (
	CommandLogWriter io.Writer
	commandLogFile   *os.File
)

func OpenCommandLogWriter(logFilePath string) error {
	if logFilePath != "" {
		outputfile, err := os.Create(logFilePath)
		if err != nil {
			return err
		}
		commandLogFile = outputfile
		CommandLogWriter = outputfile
		fmt.Println(" CommandLog writer opened with file: ", logFilePath)
	} else {
		CommandLogWriter = os.Stdout
		fmt.Println(" (!) No Command log file defined!")
		fmt.Println(" CommandLog writer opened STDOUT")
	}
	return nil
}

func WriteStringToCommandLog(s string) error {
	_, err := io.WriteString(CommandLogWriter, s)
	return err
}

func WriteLineToCommandLog(s string) error {
	return WriteStringToCommandLog(fmt.Sprintf("%s\n", s))
}

func CloseCommandLogWriter() error {
	if commandLogFile != nil {
		return commandLogFile.Close()
	}
	return nil
}
