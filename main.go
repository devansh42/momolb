package main

import (
	"flag"
	_ "github.com/google/gopacket/layers"
)

//props, contains the properties/configration of current running instance
var props properties

func setLBMethod() {
	lbmethod["roundrobin"] = 1
}

func parseArgs() {
	defer flag.Parse()
	props.islb = *flag.Bool("lb", false, "Is instance is a Load balancer")
	props.port = *flag.Uint("port", 8000, "Port to listen on")
	props.backendList = *flag.String("backend", "No default value", "backend, is the list of backends e.g <name>:<ipv4>:<port>;<name>:<ipv4>.....")
	props.lbmethod = lbmethod[*flag.String("lbmethod", "roundrobin", "Method use to load balance")]
	props.healthCheckConf = *flag.String("health", "", "method=(tcp|udp|http);port=<port>;timeout=<timeout>;interval=<interval>;threshold=<threshold>;path=<path>;httpmethod=(get|post|head|put|delete);niceStatus=<httpStatusCode>")
}

func initInstance() {
	if props.islb {
		initLB()
	} else {
		initBackend()
	}
}

func main() {
	setLBMethod()
	parseArgs()
	initInstance()
}