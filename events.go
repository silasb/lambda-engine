package main

import (
	"encoding/json"
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
			invocationRes := enqueueHandler.notifyLambda(invocation)
			nc.Publish(m.Reply, invocationRes.Res)
		}()
	})
}

type CreateFunction struct {
	FunctionName string `json:"functionName"`
	Body         string `json:"body"`
	Handler      string `json:"handler"`
}

func InitCommandControl(registerableCallback func(functionName string) error) {
	nc, _ := nats.Connect(nats.DefaultURL)
	nc.Subscribe("createFunction", func(m *nats.Msg) {
		fmt.Println("Received a message")
		nc.Publish(m.Reply, nil)

		var d CreateFunction
		json.Unmarshal(m.Data, &d)

		var envs []string
		envs = append(
			envs,
			"_HANDLER="+d.Handler,
		)

		err := uploadLambda(d.FunctionName, d.Body, envs)
		if err != nil {
			log.Println(err)
			return
		}

		registerableCallback(d.FunctionName)
	})
}
