package main

type properties struct {
	//islb, means is this instance is a load balancer of not
	islb *bool
	//port, is the port to listen on
	port *uint

	//backendString, is the string containing backend list
	backendList *string
	//healthCheckConf, is string for health configuration
	healthCheckConf *string
}

var lbmethod = make(map[string]*uint)
