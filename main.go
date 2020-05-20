package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

type Invocation struct {
	Req []byte
	Res []byte
}

type Handler struct {
	config     Config
	workCh     chan *Invocation
	responseCh chan *Invocation
	proCh      chan *Config
}

func main() {
	config := ParseConfig()

	work := make(chan *Invocation)
	responseChannel := make(chan *Invocation)
	processChannel := make(chan *Config)

	r := mux.NewRouter()

	r.HandleFunc("/2018-06-01/runtime/invocation/next", nextHandler(work))
	r.HandleFunc("/2018-06-01/runtime/invocation/{id}/response", responseHandler(responseChannel))

	ofR := mux.NewRouter()
	for domain, handlerDefinition := range config {
		fmt.Println(domain)
		enqueueHandler := Handler{handlerDefinition, work, responseChannel, processChannel}

		s := ofR.Host("{subdomain}.pyserve.com").Subrouter()
		s.PathPrefix("/").Handler(enqueueHandler)
		// ofR.PathPrefix("/").Handler(enqueueHandler)
	}

	http.Handle("/", ofR)
	http.Handle("/2018-06-01/", r)

	ofServer := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("port")),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("shim_port")),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		log.Printf("Lambda shim listening on port: %s", os.Getenv("shim_port"))
		log.Fatal(s.ListenAndServe())
		wg.Done()
	}()

	go func() {
		log.Printf("Watchdog shim listening on port: %s", os.Getenv("port"))
		log.Fatal(ofServer.ListenAndServe())
		wg.Done()
	}()

	go func() {
		log.Println("Process manager")
		processHandler(processChannel)
		wg.Done()
	}()

	ioutil.WriteFile(path.Join(os.TempDir(), ".lock"), []byte{}, 0775)

	wg.Wait()
}

func processHandler(workCh chan *Config) {
	for {
		select {
		case <-workCh:
			// fmt.Println(config)
			startProcess()
		}
	}
}

func startProcess() {
	log.Println("Starting process...")
	dns := ":9876"
	timeout := 30 * time.Second
	client, err := StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.StartProcess("blah")
	if err != nil {
		panic(err)
	}

	log.Println("Started process")
}

func stopProcess() {
	log.Println("Stopping process...")
	dns := ":9876"
	timeout := 30 * time.Second
	client, err := StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.StopProcess("blah")
	if err != nil {
		panic(err)
	}

	log.Println("Stopped process")
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("enqueue started: " + r.RequestURI)
	invocation := Invocation{}
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		log.Println("enqueue data -> " + string(body))
		invocation.Req = body
	}

	h.proCh <- &h.config
	h.workCh <- &invocation

	select {
	case invocationRes := <-h.responseCh:

		w.Write(invocationRes.Res)
		log.Println("enqueue done")
		stopProcess()
		return
	}
}

func nextHandler(workCh chan *Invocation) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		select {
		case invocation := <-workCh:

			callID := uuid.Generate().String()

			w.Header().Set("lambda-runtime-aws-request-id", callID)
			log.Println("next - " + callID)
			host, _ := os.Hostname()
			w.Header().Set("lambda-runtime-invoked-function-arn", host)

			w.WriteHeader(http.StatusOK)
			log.Println("next - [req] " + string(invocation.Req))
			w.Write(invocation.Req)

		}
	}
}

func responseHandler(responseWorkChannel chan *Invocation) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.WriteHeader(http.StatusAccepted)
		log.Println("StatusAccepted: " + r.RequestURI)

		if r.Body != nil {
			body, _ := ioutil.ReadAll(r.Body)

			log.Println("Response: " + string(body))

			invocation := Invocation{}
			invocation.Res = body
			responseWorkChannel <- &invocation
		}

	}
}
