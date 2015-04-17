package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aybabtme/tailf"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	configServerPort            = "27473"
	configOkStatusMsg           = "ok"
	configErrorStatusMsg        = "error"
	configEndOfCommandLogMarker = "_EOF__Hh2UpL4OExUSeP5LY1QaMoty97ltFqlCZaznnxjb__LOG"
	configCommandEnvPrefix      = "_CMDENV__"
)

type ResponseModel struct {
	Status   string `json:"status"`
	Msg      string `json:"msg"`
	ExitCode int    `json:"exit_code"`
}

func createErrorResponseModel(errorMessage string, exitCode int) ResponseModel {
	return ResponseModel{
		Status:   configErrorStatusMsg,
		Msg:      errorMessage,
		ExitCode: exitCode,
	}
}

func respondWithJSON(w http.ResponseWriter, respModel ResponseModel) error {
	w.Header().Set("Content Type", "application/json")
	if respModel.Status == configOkStatusMsg {
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
		Status:   configOkStatusMsg,
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
	statusMsg := configOkStatusMsg
	respMsg := "Command finished with success"
	if err != nil {
		log.Println(" [!] Error: ", err)
		WriteLineToCommandLog(fmt.Sprintf(" [!] Error: %s", err))
		statusMsg = configErrorStatusMsg
		respMsg = fmt.Sprintf("%s", err)
	}
	//
	respModel := ResponseModel{
		Status:   statusMsg,
		Msg:      respMsg,
		ExitCode: cmdExitCode,
	}

	WriteLineToCommandLog(fmt.Sprintf("%s: %s", configEndOfCommandLogMarker, statusMsg))
	WriteLineToCommandLog("-> Command Finished")

	if err := respondWithJSON(w, respModel); err != nil {
		log.Println(" [!] Failed to send Response: ", err)
	}
}

func usage() {
	fmt.Println("# Usage:")
	fmt.Println("\n## Server mode")
	fmt.Println("\nIf no parameter / flag specified it'll try to start a cmd-bridge server.")
	fmt.Println("\n## Command sender mode")
	fmt.Println("\nIf a command parameter is specified cmd-bridge will try to connect")
	fmt.Println("to an already running cmd-bridge server and execute the specified")
	fmt.Println("command through it.")
	fmt.Println("\n# Available parameters / flags:")
	fmt.Printf("\nUsage: %s [FLAGS]\n", os.Args[0])
	flag.PrintDefaults()
}

func startServer() error {
	http.HandleFunc("/cmd", commandHandler)
	http.HandleFunc("/ping", pingHandler)
	fmt.Println("Ready to serve on port:", configServerPort)
	fmt.Println()
	return http.ListenAndServe(":"+configServerPort, nil)
}

func sendJSONRequestToServer(jsonBytes []byte) error {
	resp, err := http.Post("http://localhost:27473/cmd", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		// handle error
		return err
	}
	defer resp.Body.Close()
	respBodyString := ""
	if respBodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else {
		respBodyString = string(respBodyBytes)
	}
	log.Printf("Response: %s", respBodyString)

	return nil
}

func sendCommandToServer(cmdToSend CommandModel) error {
	log.Printf("Sending command: %#v\n", cmdToSend)

	tempFile, err := makeTempFile("cmd-bridge-tmp")
	if err != nil {
		return err
	}
	tmpfilePth := tempFile.Name()
	log.Println("tmpfilePth: ", tmpfilePth)
	defer os.Remove(tmpfilePth)
	defer tempFile.Close()

	cmdToSend.LogFilePath = tmpfilePth

	cmdBytes, err := json.Marshal(cmdToSend)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		sendJSONRequestToServer(cmdBytes)
		close(done)
	}()

	isFollowFromStart := true
	follow, err := tailf.Follow(tempFile.Name(), isFollowFromStart)
	if err != nil {
		log.Fatalf("couldn't follow %q: %v", tempFile.Name(), err)
	}

	go func() {
		<-done
		if err := follow.Close(); err != nil {
			log.Fatalf("couldn't close follower: %v", err)
		}
	}()

	_, err = io.Copy(os.Stdout, follow)
	if err != nil {
		log.Fatalf("couldn't read from follower: %v", err)
		return err
	}

	return nil
}

func getCommandEnvironments() []EnvironmentKeyValue {
	cmdEnvs := []EnvironmentKeyValue{}

	for _, anEnv := range os.Environ() {
		splits := strings.Split(anEnv, "=")
		keyWithPrefix := splits[0]
		if strings.HasPrefix(keyWithPrefix, configCommandEnvPrefix) {
			cmdEnvItem := EnvironmentKeyValue{
				Key:   keyWithPrefix[len(configCommandEnvPrefix):],
				Value: os.Getenv(keyWithPrefix),
			}
			cmdEnvs = append(cmdEnvs, cmdEnvItem)
		}
	}

	log.Println("cmdEnvs: ", cmdEnvs)

	return cmdEnvs
}

func main() {
	var (
		doCommand = flag.String("do", "", "connect to a running cmd-bridge and do the specified command")
		isHelp    = flag.Bool("help", false, "show help")
	)

	flag.Usage = usage
	flag.Parse()

	if *isHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *doCommand == "" {
		fmt.Println("No command specified - starting server...")
		if err := startServer(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	doCmdEnvs := getCommandEnvironments()

	cmdToSend := CommandModel{
		Command:      *doCommand,
		Environments: doCmdEnvs,
	}
	err := sendCommandToServer(cmdToSend)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
}
