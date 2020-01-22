package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	_ "github.com/google/gopacket/layers"
)

//props, contains the properties/configration of current running instance
var props properties

func setLBMethod() {
}

func parseArgs() {
	defer glog.Info("Parsed command line arguments")

	//Due to usage of glog. CLI options for logging will be included

	props.islb = flag.Bool("lb", false, "Is instance is a Load balancer")
	props.port = flag.Uint("port", 8000, "Port to listen on")
	props.backendList = flag.String("backend", "No default value", "backend, is the list of backends e.g <name>:<ipv4>:<port>;<name>:<ipv4>.....")
	props.healthCheckConf = flag.String("health", "", "method=(tcp|udp|http);port=<port>;timeout=<timeout>;interval=<interval>;threshold=<threshold>;path=<path>;httpmethod=(get|post|head|put|delete);niceStatus=<httpStatusCode>")

	flag.Parse() //Parsing command line arguments
	fmt.Println(props.islb, props.port, props.backendList, props.healthCheckConf)
}

func initInstance() {
	if *props.islb {
		glog.Info("Load balancer is being initialized")
		initLB()
	} else {
		glog.Info("Backend is being initialized")
		initBackend()

	}
}

func main() {
	defer glog.Flush() //Flushing all the logs

	setLBMethod()
	parseArgs()
	for !flag.Parsed() {

	}
	initInstance()
}
