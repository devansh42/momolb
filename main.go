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
	props.islb = *flag.Bool("lb", false, "Is instance is a Load balancer, default is false")
	props.port = *flag.Uint("port", 8000, "Port to listen on, default is 8000")
	props.lbmethod = lbmethod[*flag.String("lbmethod", "roundrobin", "Method use to load balance, default is 'roundrobin' ")]
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
