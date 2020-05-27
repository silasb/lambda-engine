package main

import (
	"log"
	"time"

	"github.com/struCoder/pmgo/lib/master"
)

func uploadLambda(functionName string, bZipFile string, envs []string, timeout int) error {
	log.Println("Uploading lambda...")
	dns := ":9876"
	rpcTimeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, rpcTimeout)
	if err != nil {
		panic(err)
	}

	err = client.PrepareGoZip(functionName, false, nil, envs, true, bZipFile, timeout)
	if err != nil {
		return err
	}

	log.Println("lambda uploaded")

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

func restartProcess(functionName string) {
	log.Println("Restarting process...")
	dns := ":9876"
	timeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.ForceRestartProcess(functionName)
	if err != nil {
		panic(err)
	}

	log.Println("Restarted process")
}

func deleteProcess(functionName string) error {
	log.Println("Deleting function...")
	dns := ":9876"
	timeout := 5 * time.Second
	client, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	err = client.DeleteProcess(functionName)

	log.Println("Delete function...")

	return err
}

func listProcess() (*master.ProcResponse, error) {
	log.Println("List function...")
	dns := ":9876"
	timeout := 5 * time.Second
	_, err := master.StartRemoteClient(dns, timeout)
	if err != nil {
		panic(err)
	}

	process, err := getProcesses()

	log.Println("List function...")

	return process, err
}
