package main

import (
	"io/ioutil"
	"os"
)

func makeTempFile(fName string) (*os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), fName)
	if err != nil {
		return nil, err
	}
	return f, nil
}
