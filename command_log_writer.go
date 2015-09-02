package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	// CommandLogWriter ...
	CommandLogWriter io.Writer
	commandLogFile   *os.File
)

// OpenCommandLogWriter ...
func OpenCommandLogWriter(logFilePath string) error {
	if logFilePath != "" {
		outputfile, err := os.Create(logFilePath)
		if err != nil {
			return err
		}
		commandLogFile = outputfile
		CommandLogWriter = outputfile
		log.Println(" CommandLog writer opened with file: ", logFilePath)
	} else {
		CommandLogWriter = os.Stdout
		log.Println(" (!) No Command log file defined!")
		log.Println(" CommandLog writer opened STDOUT")
	}
	return nil
}

// WriteStringToCommandLog ...
func WriteStringToCommandLog(s string) error {
	_, err := io.WriteString(CommandLogWriter, s)
	return err
}

// WriteLineToCommandLog ...
func WriteLineToCommandLog(s string) error {
	return WriteStringToCommandLog(fmt.Sprintf("%s\n", s))
}

// CloseCommandLogWriter ...
func CloseCommandLogWriter() error {
	if commandLogFile != nil {
		log.Println("CommandLog file closed")
		return commandLogFile.Close()
	}
	log.Println("No CommandLog file to close")
	return nil
}
