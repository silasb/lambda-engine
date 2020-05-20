## Lambda Engine

Allowing you to run lambdas locally by emulating the lambda request/response cycle.

This is highly experimental / early-work and only a PoC.

Big thanks from https://github.com/alexellis/lambda-on-openfaas-poc for the original POC.

The big difference is that this is using a customized version of [pmgo](https://github.com/struCoder/pmgo/tree/beta) for a process manager.

The big POS here is if I can do self-hosted lambda execution with a secure runtime like Deno.

## Run locally

### Run the Lambda runtime

You'll need Node 8, 10 or 11 installed on your local machine.

```
git clone https://github.com/silasb/deno-aws-lambda-example
LAMBDA_TASK_ROOT=$PWD _HANDLER="function.handler" AWS_LAMBDA_RUNTIME_API=127.0.0.1:8081 ./bootstrap
```

### Run the lambda-engine

You'll need Go 1.13 installed locally for this part.

* Clone the shim and run it:

```
git clone https://github.com/silasb/lambda-engine
cd lambda-engine

shim_port=8081 port=8082 go run main.go config.go process.go
```

### Invoke the Lambda function

```
for i in {0..100} ; do  curl localhost:8082 -d '{"invocation": "#: '$i'"}' && echo ; done
```

If you like run this in multiple windows at the same time:

```
(
    for i in {0..100} ; do  curl localhost:8082 -d '{"invocation": "#: '$i'"}' && echo ; done &
    for i in {101..201} ; do  curl localhost:8082 -d '{"invocation": "#: '$i'"}' && echo ; done &
    for i in {202..302} ; do  curl localhost:8082 -d '{"invocation": "#: '$i'"}' && echo ; done &
)
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
