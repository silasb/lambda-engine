package main

import (
	"log"
	"time"

	"github.com/struCoder/pmgo/lib/master"
)

func uploadLambda(functionName string, bZipFile string, envs []string) error {
	log.Println("Uploading lambda...")
	dns := ":9876"
	timeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.PrepareGoZip(functionName, false, nil, envs, true, bZipFile)
	if err != nil {
		return err
	}

	log.Println("Started process")

	return nil
}

func startProcessEnvs(procName string, envs []string) {
	log.Println("Starting process...")
	dns := ":9876"
	timeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.StartProcessEnvs(procName, envs)
	if err != nil {
		panic(err)
	}

	log.Println("Started process")
}

func getProcesses() (*master.ProcResponse, error) {
	log.Println("Getting processes...")
	dns := ":9876"
	timeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		return nil, err
	}

	processes, err := client.MonitStatus()

	// log.Printf("Got processes: %s", processes)
	return &processes, err
}

/*
func stopProcess() {
	log.Println("Stopping process...")
	dns := ":9876"
	timeout := 5 * time.Second
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
*/
