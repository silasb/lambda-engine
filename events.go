package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func InitNatsConsumer(functionName string, enqueueHandler Handler) {
	setupNats(functionName, enqueueHandler)
}

func setupNats(functionName string, enqueueHandler Handler) {
	fmt.Println("Configuring: ", functionName)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	// defer nc.Close()

	nc.Subscribe(functionName, func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
		invocation := Invocation{}
		if len(m.Data) == 0 {
			invocation.Req = []byte("{}")
		} else {
			invocation.Req = m.Data
		}
		log.Println("enqueue data -> " + string(invocation.Req))

		go func() {
			invocationRes, err := enqueueHandler.notifyLambda(invocation)
			if err != nil {
				nc.Publish(m.Reply, []byte(err.Error()))
			} else {
				nc.Publish(m.Reply, invocationRes.Res)
			}
		}()
	})
}

type CreateFunction struct {
	FunctionName string `json:"functionName"`
	Body         string `json:"body"`
	Handler      string `json:"handler"`
	Timeout      int    `json:"timeout"`
}

func InitCommandControl(registerableCallback RegisterableCb) {
	nc, _ := nats.Connect(nats.DefaultURL)
	c, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)

	c.Subscribe("createFunction", func(d *CreateFunction, reply string) {
		//var d CreateFunction
		//json.Unmarshal(m.Data, &d)
		fmt.Printf("Received a message: %+v\n", d)
		c.Publish(reply, nil)

		var envs []string
		envs = append(
			envs,
			"_HANDLER="+d.Handler,
		)

		err := uploadLambda(d.FunctionName, d.Body, envs, d.Timeout)
		if err != nil {
			log.Println(err)
			return
		}

		registerableCallback(d.FunctionName, d.Timeout)
	})

	c.Subscribe("deleteFunction", func(subj string, reply string, msg string) {
		fmt.Printf("Received a message: %+v %+v\n", subj, msg)

		err := deleteProcess(msg)
		if err != nil {
			c.Publish(reply, []byte(err.Error()))
		} else {
			c.Publish(reply, nil)
		}
	})

	c.Subscribe("listFunctions", func(subj string, reply string, msg string) {
		fmt.Printf("Received a message: %+v\n", subj)

		processes, err := listProcess()
		if err != nil {
			c.Publish(reply, []byte(err.Error()))
		} else {
			c.Publish(reply, processes)
		}
	})
}
