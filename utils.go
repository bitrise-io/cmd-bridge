package main

import (
	"io/ioutil"
	"log"
	"os"
)

func makeTempFile(fName string) (*os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), fName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func vLogln(s string, args ...interface{}) {
	if Config_IsVerboseLogMode {
		log.Println(s, args)
	}
}

func vLogf(s string, args ...interface{}) {
	if Config_IsVerboseLogMode {
		log.Printf(s, args)
	}
}
