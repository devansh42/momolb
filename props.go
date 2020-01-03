package main

type properties struct {
	//islb, means is this instance is a load balancer of not
	islb bool
	//port, is the port to listen on
	port uint
	//lbmethod, is id of lb method used
	lbmethod uint
}

var lbmethod = make(map[string]uint)
