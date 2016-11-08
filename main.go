package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hpcloud/tail"
	_ "golang.org/x/sys/unix"
)

var (
	configServerPort       = "27473"
	configOkStatusMsg      = "ok"
	configErrorStatusMsg   = "error"
	configCommandEnvPrefix = "_CMDENV__"
)

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

// ResponseModel ...
type ResponseModel struct {
	Status   string `json:"status"`
	Msg      string `json:"msg"`
	ExitCode int    `json:"exit_code"`
}

//
// --- Server mode

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

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(" [!] Failed to close r.Body:", err)
		}
	}()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp := createErrorResponseModel(
			fmt.Sprintf("Failed to ready Request Body: %s", err),
			1,
		)
		if err := respondWithJSON(w, resp); err != nil {
			log.Printf("Failed to respond with JSON: %#v", resp)
		}
		return
	}

	bodyString := string(bodyBytes)
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
		if err := respondWithJSON(w, resp); err != nil {
			log.Printf("Failed to respond with JSON: %#v", resp)
		}
		return
	}
	fmt.Printf("Command to run: %#v\n", cmdToRun)

	err = OpenCommandLogWriter(cmdToRun.LogFilePath)
	cmdExitCode := 0
	if err == nil {
		defer func() {
			if err := CloseCommandLogWriter(); err != nil {
				log.Println(" [!] Failed to CloseCommandLogWriter:", err)
			}
		}()
		cmdExitCode, err = ExecuteCommand(cmdToRun)
	}

	//
	// Response
	statusMsg := configOkStatusMsg
	respMsg := "Command finished with success"
	if err != nil {
		log.Println(" [!] Error: ", err)
		// WriteLineToCommandLog(fmt.Sprintf(" [!] Error: %s", err))
		statusMsg = configErrorStatusMsg
		respMsg = fmt.Sprintf("%s", err)
	}
	//
	respModel := ResponseModel{
		Status:   statusMsg,
		Msg:      respMsg,
		ExitCode: cmdExitCode,
	}

	if ConfigIsVerboseLogMode {
		if err := WriteLineToCommandLog("-> Command Finished"); err != nil {
			log.Println(" [!] Failed to write 'Command Finished' into Command Log")
		}
	}

	if err := respondWithJSON(w, respModel); err != nil {
		log.Println(" [!] Failed to send Response: ", err)
	}
}

func startServer() error {
	http.HandleFunc("/cmd", commandHandler)
	http.HandleFunc("/ping", pingHandler)
	fmt.Println("Ready to serve on port:", configServerPort)
	fmt.Println()
	return http.ListenAndServe(":"+configServerPort, nil)
}

//
// --- non server mode

func sendJSONRequestToServer(jsonBytes []byte) (cmdExCode int, cmdErr error) {
	cmdExCode = 1
	cmdErr = nil

	resp, err := http.Post("http://localhost:27473/cmd", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		// handle error
		log.Println("Failed to send command to cmd-bridge server: ", err)
		return 1, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(" [!] Failed to close resp.Body:", err)
		}
	}()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read cmd-bridge server response: ", err)
		return 1, err
	}
	respBodyString := string(respBodyBytes)

	vLogln("Response: ", respBodyString)

	var respModel ResponseModel
	jsonParser := json.NewDecoder(strings.NewReader(respBodyString))
	if err := jsonParser.Decode(&respModel); err != nil {
		log.Println("Failed to decode cmd-bridge server response (JSON): ", err)
		return 1, err
	}
	vLogf("respModel: %#v\n", respModel)
	cmdExCode = respModel.ExitCode

	if respModel.Status != configOkStatusMsg {
		return cmdExCode, fmt.Errorf("Server returned an error response: %#v", respModel)
	}

	if respModel.ExitCode != 0 {
		return cmdExCode, fmt.Errorf("Bridged command exit code is not 0: %#v", respModel)
	}

	return cmdExCode, nil
}

func sendCommandToServer(cmdToSend CommandModel, isVerbose bool) (cmdExCode int, cmdErr error) {
	cmdExCode = 1
	cmdErr = nil

	tempFile, err := makeTempFile("cmd-bridge-tmp")
	if err != nil {
		return 1, err
	}
	tmpfilePth := tempFile.Name()
	vLogln("tmpfilePth: ", tmpfilePth)
	defer func() {
		if err := os.Remove(tmpfilePth); err != nil {
			log.Println(" [!] Failed to os.Remove(tmpfilePth):", err)
		}
		if err := tempFile.Close(); err != nil {
			log.Println(" [!] Failed to tempFile.Close:", err)
		}
	}()

	cmdToSend.LogFilePath = tmpfilePth

	vLogln(fmt.Sprintf("Sending command: %#v", cmdToSend))

	cmdBytes, err := json.Marshal(cmdToSend)
	if err != nil {
		return 1, err
	}

	t, err := tail.TailFile(tempFile.Name(), tail.Config{Follow: true})
	if err != nil {
		return 1, err
	}
	go func() {
		cmdExCode, cmdErr = sendJSONRequestToServer(cmdBytes)
		if err := t.Stop(); err != nil {
			log.Println(" [!] Failed to (tail) t.Stop():", err)
		}
	}()

	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	return cmdExCode, cmdErr
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

	vLogf("cmdEnvs: %#v\n", cmdEnvs)

	return cmdEnvs
}

func main() {
	var (
		doCommand      = flag.String("do", "", "Connect to a running cmd-bridge and do the specified command")
		flagCmdWorkDir = flag.String("workdir", "", "Working directory of the specified command.")
		isHelp         = flag.Bool("help", false, "Show help")
		isVerbose      = flag.Bool("verbose", false, "Verbose output")
		isVersion      = flag.Bool("version", false, "Prints version")
	)

	flag.Usage = usage
	flag.Parse()

	if *isHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *isVersion {
		fmt.Println(VersionString)
		os.Exit(0)
	}

	if *isVerbose == true {
		log.Println(" (i) Verbose mode")
		ConfigIsVerboseLogMode = true
	}

	// --- server mode

	if *doCommand == "" {
		fmt.Println("No command specified - starting server...")
		if err := startServer(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// --- non-server mode

	doCmdEnvs := getCommandEnvironments()
	cmdToSend := CommandModel{
		Command:          *doCommand,
		Environments:     doCmdEnvs,
		WorkingDirectory: *flagCmdWorkDir,
	}
	cmdExCode, cmdErr := sendCommandToServer(cmdToSend, *isVerbose)
	if cmdErr != nil {
		vLogln("Error: ", cmdErr)
		if cmdExCode != 0 {
			os.Exit(cmdExCode)
		}
		vLogln("Command returned an exit code 0 and an error - we'll return an exit code 1")
		os.Exit(1)
	}
	if cmdExCode != 0 {
		vLogln("No error returned, but command exit code was:", cmdExCode)
		os.Exit(cmdExCode)
	}
	os.Exit(0)
}
