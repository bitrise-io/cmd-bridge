package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
)

var (
	serverPort            = "27473"
	okStatusMsg           = "ok"
	errorStatusMsg        = "error"
	endOfCommandLogMarker = "_EOF__Hh2UpL4OExUSeP5LY1QaMoty97ltFqlCZaznnxjb__LOG"
)

type ResponseModel struct {
	Status   string `json:"status"`
	Msg      string `json:"msg"`
	ExitCode int    `json:"exit_code"`
}

func createErrorResponseModel(errorMessage string, exitCode int) ResponseModel {
	return ResponseModel{
		Status:   errorStatusMsg,
		Msg:      errorMessage,
		ExitCode: exitCode,
	}
}

func respondWithJSON(w http.ResponseWriter, respModel ResponseModel) error {
	w.Header().Set("Content Type", "application/json")
	if respModel.Status == okStatusMsg {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	log.Printf("=> Response: %#v\n", respModel)

	err := json.NewEncoder(w).Encode(&respModel)
	return err
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(" (i) Ping received")

	//
	respModel := ResponseModel{
		Status:   okStatusMsg,
		Msg:      "pong",
		ExitCode: 0,
	}

	if err := respondWithJSON(w, respModel); err != nil {
		log.Println(" [!] Failed to send Response: ", err)
	}
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(" (i) Command received")

	defer r.Body.Close()
	bodyString := ""
	if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
		resp := createErrorResponseModel(
			fmt.Sprintf("Failed to ready Request Body: %s", err),
			1,
		)
		respondWithJSON(w, resp)
		return
	} else {
		bodyString = string(bodyBytes)
	}
	log.Println(" (i) Raw request body: ", bodyString)

	// queryValues := r.URL.Query()
	// decoder := json.NewDecoder(r.Body)
	decoder := json.NewDecoder(strings.NewReader(bodyString))
	var cmdToRun CommandModel
	if err := decoder.Decode(&cmdToRun); err != nil {
		resp := createErrorResponseModel(
			fmt.Sprintf("Invalid JSON: %s", err),
			1,
		)
		respondWithJSON(w, resp)
		return
	}
	fmt.Printf("Command to run: %#v\n", cmdToRun)

	err := OpenCommandLogWriter(cmdToRun.LogFilePath)
	cmdExitCode := 0
	if err == nil {
		defer CloseCommandLogWriter()

		// WriteLineToCommandLog(fmt.Sprintf(" (i) Using Command Params: %#v", commandParams))
		// err = commandParams.Validate()
		// if err == nil {
		err = ExecuteCommand(cmdToRun)
		if err != nil {
			// Did the command fail because of an unsuccessful exit code
			var waitStatus syscall.WaitStatus
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				exCode := waitStatus.ExitStatus()
				fmt.Println("Exit status: ", exCode)
				cmdExitCode = exCode
			}
		}
		// }
	}

	//
	// Response
	statusMsg := okStatusMsg
	respMsg := "Command finished with success"
	if err != nil {
		log.Println(" [!] Error: ", err)
		WriteLineToCommandLog(fmt.Sprintf(" [!] Error: %s", err))
		statusMsg = errorStatusMsg
		respMsg = fmt.Sprintf("%s", err)
	}
	//
	respModel := ResponseModel{
		Status:   statusMsg,
		Msg:      respMsg,
		ExitCode: cmdExitCode,
	}

	WriteLineToCommandLog(fmt.Sprintf("%s: %s", endOfCommandLogMarker, statusMsg))
	WriteLineToCommandLog("-> Command Finished")

	if err := respondWithJSON(w, respModel); err != nil {
		log.Println(" [!] Failed to send Response: ", err)
	}
}

func main() {
	http.HandleFunc("/cmd", commandHandler)
	http.HandleFunc("/ping", pingHandler)
	fmt.Println("Ready to serve on port:", serverPort)
	fmt.Println()
	http.ListenAndServe(":"+serverPort, nil)
}
