## Lambda Engine

Allowing you to run lambdas locally by emulating the lambda request/response cycle.

This is highly experimental / early-work and only a PoC.

Big thanks from https://github.com/alexellis/lambda-on-openfaas-poc for the original POC.

The big difference is that this is using a customized version of [pmgo](https://github.com/struCoder/pmgo/tree/beta) for a process manager.

The big POS here is if I can do self-hosted lambda execution with a secure runtime like Deno.

## Run locally

### Run Nats

```
docker run -d --name nats-main -p 4222:4222 -p 6222:6222 -p 8222:8222 nats
```

### Run the lambda-process-manager

```
git clone https://github.com/silasb/lambda-scheduler
cd lambda-scheduler
go build
./pmgo serve
```

### Run the lambda-engine

You'll need Go 1.13 installed locally for this part.

* Clone the shim and run it:

```
git clone https://github.com/silasb/lambda-engine
cd lambda-engine

fd .go | entr -c -r -s "shim_port=8081 port=8082 go run main.go process.go events.go"
```

### Create the Lambda functions

```
go run cli.go create-function --function-name h1 --zip-file fileb://examples/function.zip --handler function.handler1
go run cli.go create-function --function-name h2 --zip-file fileb://examples/function.zip --handler function.handler2
go run cli.go create-function --function-name h3 --zip-file fileb://examples/function.zip --handler function.handler3
```

### Invoke the Lambda function via NATS

```
❯ go run cli.go invoke --function-name h1 --payload '{"hello": "hi"}'
2020/05/21 22:30:36 h1 {"hello": "hi"}
{"statusCode":200,"body":"{\"version\":{\"deno\":\"1.0.0\",\"v8\":\"8.4.300\",\"typescript\":\"3.9.2\"},\"build\":{\"target\":\"x86_64-unknown-linux-gnu\",\"arch\":\"x86_64\",\"os\":\"linux\",\"vendor\":\"unknown\",\"env\":\"gnu\"},\"event\":{\"hello\":\"hi\"}}"}
❯ go run cli.go invoke --function-name h2 --payload '{"hello": "hi"}'
2020/05/21 22:30:39 h2 {"hello": "hi"}
{"statusCode":200,"body":"{\"hello\":\"world\"}"}
❯ go run cli.go invoke --function-name h3 --payload '{"hello": "hi"}'
2020/05/21 22:30:53 h3 {"hello": "hi"}
"hello world"
```

### Invoke the Lambda function via HTTP

```
❯ curl localhost:8082/test -H 'Host: h1.pyserve.com' -d '{"hello": "world"}'
{"version":{"deno":"1.0.0","v8":"8.4.300","typescript":"3.9.2"},"build":{"target":"x86_64-unknown-linux-gnu","arch":"x86_64","os":"linux","vendor":"unknown","env":"gnu"},"event":{"hello":"world"}}
❯ curl localhost:8082/test -H 'Host: h2.pyserve.com' -d '{"hello": "world"}'
{"hello":"world"}
```

## Conceptual diagram

![](./concept.png)

## How do we know it works?

You can test it out using the Node.js tester program in the verify folder.

It verifies that if a request inputs a certain number that it's also echoed in the response.

```
cd verify
npm i
node index.js
```

If all the responses are correct you'll see `[x]` printed back on each line, otherwise the delta.

Test in parallel:

```
(
    node index.js &
    node index.js &
    node index.js &
)
```

## What next?

I have no idea, this is just a POC.

See also:

* [Lambda custom runtimes](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html)
* [Lambda Runtime Interface](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html)
* [Blog: OpenFaaS Template Store](https://www.openfaas.com/blog/template-store/)
