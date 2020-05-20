package main

import (
	"net"
	"net/rpc"
	"time"
)

type RemoteClient struct {
	conn *rpc.Client // RpcConnection for the remote client.
}

func StartRemoteClient(dsn string, timeout time.Duration) (*RemoteClient, error) {
	conn, err := net.DialTimeout("tcp", dsn, timeout)
	if err != nil {
		return nil, err
	}
	return &RemoteClient{conn: rpc.NewClient(conn)}, nil
}

func (client *RemoteClient) StartProcess(procName string) error {
	var started bool
	return client.conn.Call("RemoteMaster.StartProcess", procName, &started)
}

func (client *RemoteClient) StopProcess(procName string) error {
	var stopped bool
	return client.conn.Call("RemoteMaster.StopProcess", procName, &stopped)
}

// client, err := master.StartRemoteClient(dsn, timeout)
