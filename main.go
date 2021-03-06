package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

type Invocation struct {
	Req        []byte
	Res        []byte
	responseCh chan *Invocation
	callID     string
}

type LambdaResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

type Handler struct {
	Name    string
	workCh  chan *Invocation
	Timeout int
}

func registerLambda(functionName string, timeout int) error {
	workCh := make(WorkCh)
	responseCh := make(ResponseCh)

	handler := Handler{functionName, workCh, timeout}

	registerNats(functionName, handler)
	port := registerLambdaWeb(workCh, responseCh, handler)
	envs := []string{fmt.Sprintf("AWS_LAMBDA_RUNTIME_API=127.0.0.1:%d", port)}
	startProcessEnvs(functionName, envs)

	registerPublicWeb(functionName, handler)

	log.Printf("Lambda shim for %s listening on port: %d\n", functionName, port)
	return nil
}

func registerNats(functionName string, handler Handler) {
	InitNatsConsumer(functionName, handler)
}

func registerLambdaWeb(workCh WorkCh, responseCh ResponseCh, handler Handler) int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/2018-06-01/runtime/invocation/next", nextHandler(workCh))
	r.HandleFunc("/2018-06-01/runtime/invocation/{callID}/response", responseHandler())

	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", "0"),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        r,
	}

	port := listener.Addr().(*net.TCPAddr).Port

	wg.Add(1)
	go func() {
		log.Fatal(server.Serve(listener))
		wg.Done()
	}()

	return port
}

func registerPublicWeb(functionName string, handler Handler) error {
	s := upstreamMux.Host(fmt.Sprintf("%s.pyserve.com", functionName)).Subrouter()
	s.PathPrefix("/").Handler(handler)

	return nil
}

type RegisterableCb func(functionName string, timeout int) error
type WorkCh chan *Invocation
type ResponseCh chan *Invocation

var wg sync.WaitGroup
var upstreamMux *mux.Router
var responseMapper map[string]*Invocation

func main() {
	responseMapper = make(map[string]*Invocation, 10)
	wg = sync.WaitGroup{}

	upstreamMux = mux.NewRouter()

	InitCommandControl(registerLambda)

	processes, _ := getProcesses()
	for _, process := range processes.Procs {
		fmt.Printf("%+v\n", process.Name)
		registerLambda(process.Name, process.Timeout)
	}

	// http.Handle("/", ofR)
	// http.Handle("/2018-06-01/", r)

	upstreamServer := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("port")),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        upstreamMux,
	}

	wg.Add(1)
	go func() {
		log.Printf("Upstream server on port: %s", os.Getenv("port"))
		log.Fatal(upstreamServer.ListenAndServe())
		wg.Done()
	}()

	ioutil.WriteFile(path.Join(os.TempDir(), ".lock"), []byte{}, 0775)

	wg.Wait()
}

func (h Handler) notifyLambda(invocation Invocation) (*Invocation, error) {
	fmt.Printf("notifyLambda: %+v\n", h)

	callID := uuid.Generate().String()
	responseCh := make(ResponseCh)
	invocation.responseCh = responseCh
	responseMapper[callID] = &invocation

	invocation.callID = callID
	h.workCh <- &invocation

	select {
	case invocationRes := <-invocation.responseCh:
		return invocationRes, nil

	case <-time.After(time.Duration(h.Timeout) * time.Second):
		// case <-ctx.Done():
		err := errors.New("Lambda timeout occurred")
		log.Println(err)

		// not safe
		delete(responseMapper, callID)

		return &Invocation{}, err
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("enqueue started: " + r.RequestURI)
	invocation := Invocation{}
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) == 0 {
			invocation.Req = []byte("{}")
		} else {
			invocation.Req = body
		}
		log.Println("enqueue data -> " + string(invocation.Req))
	}

	invocationRes, _ := h.notifyLambda(invocation)
	var res LambdaResponse
	json.Unmarshal(invocationRes.Res, &res)

	w.WriteHeader(res.StatusCode)
	io.WriteString(w, res.Body)
	log.Println("enqueue done")
	return
}

func nextHandler(workCh chan *Invocation) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		select {
		case invocation := <-workCh:

			w.Header().Set("lambda-runtime-aws-request-id", invocation.callID)
			log.Println("next - " + invocation.callID)
			host, _ := os.Hostname()
			w.Header().Set("lambda-runtime-invoked-function-arn", host)

			w.WriteHeader(http.StatusOK)
			log.Println("next - [req] " + string(invocation.Req))
			w.Write(invocation.Req)
		}
	}
}

func responseHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		vars := mux.Vars(r)

		w.WriteHeader(http.StatusAccepted)
		log.Println("StatusAccepted: " + r.RequestURI)

		if r.Body != nil {
			invocation, ok := responseMapper[vars["callID"]]
			if ok {
				body, _ := ioutil.ReadAll(r.Body)

				log.Printf("Response %s: %s\n", vars["callID"], string(body))
				invocation.Res = body
				invocation.responseCh <- invocation
				// close(invocation.responseCh)
			} else {
				log.Printf("Not accepting response since timeout occurred")
			}
		}

	}
}
