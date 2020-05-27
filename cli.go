package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/struCoder/pmgo/lib/master"
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
	createTimeout      = createFunction.Flag("timeout", "").Int()

	deleteFunctionCmd  = app.Command("delete-function", "delete an function")
	deleteFunctionName = deleteFunctionCmd.Flag("function-name", "").Required().String()

	listFunctionsCmd = app.Command("list-functions", "list functions")

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

func client() *nats.EncodedConn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		panic(err)
	}

	return c
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case invoke.FullCommand():
		log.Println(*invokeFunctionName, *invokePayload)

		c := client()

		// err = nc.Publish(*functionName, []byte(*payload))
		var msg string
		err := c.Request(*invokeFunctionName, *invokePayload, &msg, 901*time.Second)
		if err != nil {
			panic(err)
		}

		fmt.Println(msg)

		c.Close()

	case createFunction.FullCommand():
		log.Println(*createFunctionName, *createZipFile, *createHandler)
		buildFunction()
	case deleteFunctionCmd.FullCommand():
		log.Println(*deleteFunctionName)
		deleteFunction()
	case listFunctionsCmd.FullCommand():
		listFunctions()
	}
}

type CreateFunction struct {
	FunctionName string `json:"functionName"`
	Body         string `json:"body"`
	Handler      string `json:"handler"`
	Timeout      int    `json:"timeout"`
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

		var timeout int
		if *createTimeout == 0 {
			timeout = 3
		} else {
			timeout = *createTimeout
		}

		payload := CreateFunction{
			FunctionName: *createFunctionName,
			Body:         encoded,
			Handler:      *createHandler,
			Timeout:      timeout,
		}

		c := client()

		var msg string
		err = c.Request("createFunction", payload, &msg, 1*time.Second)
		if err != nil {
			panic(err)
		}

		fmt.Println(msg)

		c.Close()

	} else {
		panic("unknown file scheme")
	}
}

func deleteFunction() {
	c := client()

	var msg string
	err := c.Request("deleteFunction", *deleteFunctionName, &msg, 5*time.Second)
	if err != nil {
		panic(err)
	}

	fmt.Println(msg)

	c.Close()
}

func listFunctions() {
	c := client()

	var processes *master.ProcResponse
	err := c.Request("listFunctions", nil, &processes, 5*time.Second)
	if err != nil {
		panic(err)
	}

	for _, process := range processes.Procs {
		fmt.Printf("%+v\n", process.Name)
	}

	c.Close()
}
