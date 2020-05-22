module lambda-on-openfass-poc

go 1.13

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/nats-io/nats-server/v2 v2.1.7 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/struCoder/pmgo v0.5.2-0.20200103011450-c3568922e94f
	github.com/zclconf/go-cty v1.4.1 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
)

// replace github.com/struCoder/pmgo/lib/master => github.com/silasb/lambda-scheduler/lib/master beta

replace github.com/struCoder/pmgo => ../lambda-scheduler
