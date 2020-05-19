package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

type Invocation struct {
	Req []byte
	Res []byte
}

type Handler struct {
	workCh     chan *Invocation
	responseCh chan *Invocation
	proCh      chan string
}

func main() {

	work := make(chan *Invocation)
	responseChannel := make(chan *Invocation)
	processChannel := make(chan string)

	r := mux.NewRouter()

	r.HandleFunc("/2018-06-01/runtime/invocation/next", nextHandler(work))
	r.HandleFunc("/2018-06-01/runtime/invocation/{id}/response", responseHandler(responseChannel))

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("shim_port")),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	ofR := mux.NewRouter()
	enqueueHandler := Handler{work, responseChannel, processChannel}
	ofR.PathPrefix("/").Handler(enqueueHandler)

	http.Handle("/", ofR)
	http.Handle("/2018-06-01/", r)
	ofServer := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("port")),
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
		// log.Fatal(ofServer.ListenAndServe())
		wg.Done()
	}()

	ioutil.WriteFile(path.Join(os.TempDir(), ".lock"), []byte{}, 0775)

	wg.Wait()
}

func processHandler(workCh chan string) {
	select {
	case b := <-workCh:
		fmt.Println(b)
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("enqueue started: " + r.RequestURI)
	invocation := Invocation{}
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		log.Println("enqueue data -> " + string(body))
		invocation.Req = body
	}

	h.proCh <- "Start"
	h.workCh <- &invocation

	select {
	case invocationRes := <-h.responseCh:

		w.Write(invocationRes.Res)
		log.Println("enqueue done")
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
