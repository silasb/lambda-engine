package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("invoke", "invoke")
	// dns     = app.Flag("dns", "TCP Dns host.").Default(":9876").String()
	// timeout = 30 * time.Second

	invoke             = app.Command("invoke", "invoke")
	invokeFunctionName = invoke.Flag("function-name", "").Required().String()
	invokePayload      = invoke.Flag("payload", "").String()

	createFunction     = app.Command("create-function", "create a function")
	createFunctionName = createFunction.Flag("function-name", "").Required().String()
	createZipFile      = createFunction.Flag("zip-file", "").Required().String()
	createHandler      = createFunction.Flag("handler", "").Required().String()
	// serveConfigFile = serve.Flag("config-file", "Config file location").String()

	// resurrect = app.Command("resurrect", "Resurrect all previously save processes.")

	// start           = app.Command("start", "start and daemonize an app.")
	// startSourcePath = start.Arg("start go file", "go file.").Required().String()
	// startName       = start.Arg("name", "Process name.").Required().String()
	// binFile         = start.Arg("binary", "compiled golang file").Bool()
	// startKeepAlive  = true
	// startArgs       = start.Flag("args", "External args.").Strings()
	// startEnvs       = start.Flag("envs", "External envs.").Strings()

	// restart     = app.Command("restart", "Restart a process.")
	// restartName = restart.Arg("name", "Process name.").Required().String()

	// stop     = app.Command("stop", "Stop a process.")
	// stopName = stop.Arg("name", "Process name.").Required().String()

	// delete     = app.Command("delete", "Delete a process.")
	// deleteName = delete.Arg("name", "Process name.").Required().String()

	// save = app.Command("save", "Save a list of processes onto a file.")

	// status = app.Command("list", "Get pmgo list.")

	// version        = app.Command("version", "get version")
	// currentVersion = "0.6.0"

	// info     = app.Command("info", "Describe importance parameters of a process id")
	// infoName = info.Arg("name", "process name").Required().String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case invoke.FullCommand():
		log.Println(*invokeFunctionName, *invokePayload)

		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		// err = nc.Publish(*functionName, []byte(*payload))
		msg, err := nc.Request(*invokeFunctionName, []byte(*invokePayload), 1*time.Second)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(msg.Data))

		nc.Close()

	case createFunction.FullCommand():
		log.Println(*createFunctionName, *createZipFile, *createHandler)
		buildFunction()
	}
}

func buildFunction() {
	url, err := url.Parse(*createZipFile)
	if err != nil {
		panic(err)
	}

	if url.Scheme == "fileb" {
		file, err := os.Open(url.Host)
		if err != nil {
			panic(err)
		}

		b, _ := ioutil.ReadAll(file)
		encoded := base64.StdEncoding.EncodeToString(b)

		c := struct {
			FunctionName string `json:"functionName"`
			Body         string `json:"body"`
			Handler      string `json:"handler"`
		}{
			FunctionName: *createFunctionName,
			Body:         encoded,
			Handler:      *createHandler,
		}

		payload, err := json.Marshal(&c)
		if err != nil {
			panic(err)
		}

		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		msg, err := nc.Request("createFunction", payload, 1*time.Second)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(msg.Data))

		nc.Close()

	} else {
		panic("unknown file scheme")
	}

}
