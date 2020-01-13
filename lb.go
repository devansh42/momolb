package main

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"stathat.com/c/consistent"
)

const (
	IPV4 = "ipv4"
)

var lbincomingPacket = make(chan *layers.IPv4)

//initLB, initiates Load Balancer
func initLB() {
	go intializeHealthChecker()
	go intializeBackend()
	go handleLBIngress()
	go healthCheckService()
	go manageBackEnd()
	go packetSenderListner()

}

//intializeHealthChecker, parse health checker configuration string
func intializeHealthChecker() {
	//validParams := map[string]bool{}

	var healthChecker instanceHealthChecker
	validParams := make(map[string]string)
	conf := strings.Split(props.healthCheckConf, ";")
	for _, prop := range conf {
		kv := strings.Split(prop, "=")
		if len(kv) != 2 {
			continue //invalid prop
		}
		validParams[kv[0]] = kv[1]
	}
	//assuming all the things are working
	//suppressing input validation
	port, err := strconv.ParseUint(validParams["port"], 10, 32)
	if err != nil {
		port = 80
		return
	}
	thres, err := strconv.ParseFloat(validParams["threshold"], 32)
	if err != nil {
		thres = 0.5
	}

	timeout, err := time.ParseDuration(validParams["timeout"])
	if err != nil {
		timeout = time.Second * 5
	}
	interval, err := time.ParseDuration(validParams["interval"])
	if err != nil {
		interval = time.Second * 30
	}
	switch validParams["method"] {

	case "tcp":
	case "udp":
		x := tcporudphealthcheck{}
		x.port = uint(port)
		x.timeout = timeout
		x.interval = interval
		x.threshold = float32(thres)
		if validParams["method"] == "udp" {
			x.protocol = 1
		} else {
			x.protocol = 0
		}
		healthChecker = x
	case "http":
		x := httphealthchecker{}
		x.port = uint(port)
		x.timeout = timeout
		x.interval = interval
		x.threshold = float32(thres)
		if httpmethod, ok := validParams["httpmethod"]; ok {
			x.method = httpmethod
		} else {
			x.method = "get" //default request method
		}
		niceStatus, err := strconv.ParseInt(validParams["niceStatus"], 10, 16)
		if err != nil {
			niceStatus = 200 //default http response
		}
		path, ok := validParams["path"]
		if ok {
			x.path = path
		} else {
			x.path = "/index.html" //default path
		}
		x.niceStatus = int(niceStatus)
		healthChecker = x
	default:
	}

	globalHealthCheckerCh <- healthChecker //setting global health checker
	close(globalHealthCheckerCh)           //closing channel just for one time

}

//intializeBackend, initializes backend list as parsed in the argument
//Format is <name>:<ipv4>:<port>;<name>:<ipv4>:......
func intializeBackend() {
	list := props.backendList
	bl := strings.Split(list, ";") //to split list of backend
	pool := make([]backend, len(bl))
	i := 0
	for _, back := range bl {
		bb := strings.Split(back, ":")
		if len(bb) != 3 {
			continue //Invalid backend name
		}
		port, err := strconv.ParseUint(bb[2], 10, 16)
		if err != nil {
			continue //Invalid Port
		}
		ip := net.ParseIP(bb[1])
		if ip == nil {
			continue //invalid ip
		}
		pool[i] = backend{Name: bb[0], IP: ip, Port: uint16(port)}
		i++
	}
	if len(pool) == 0 {
		//no valid backend
		panic(errors.New("Couldn't initalize backend : No valid configuration found"))
	}

	bp := new(backendPool)
	bp.pool = pool
	bp.healthChecker = <-globalHealthCheckerCh //waiting for health check pass
	bp.consistency = consistent.New()          //new consistency hasher
	globalbackendPool <- bp                    //sending global backend pool
	close(globalbackendPool)                   //closing the backend pool channel
}

//handleLBIngress, handles ingress traffic for load balancer
func handleLBIngress() {
	//Taking a incoming ip packet coming from given port endpoint

	for x := range lbincomingPacket {
		//Now we have a tcp packet let's distribute over the network
		xip := gopacket.NewPacket(x.LayerContents(), layers.LayerTypeIPv4, gopacket.Default)
		tcplayer := xip.TransportLayer()
		var srport layers.TCPPort
		if tcplayer != nil {
			if tcp, ok := tcplayer.(*layers.TCP); ok {
				srport = tcp.SrcPort
			}

		} else {
			continue
		}
		ip := &layers.IPv4{
			SrcIP:    x.DstIP,
			Protocol: 0x2f,                         //47
			TTL:      0xff,                         //255 secs
			DstIP:    nextBackEnd(x.SrcIP, srport), //returning next backend to get request
		}
		gre := &layers.GRE{
			Protocol: 0x0800, //As encapsulating packet is IP packet
		}
		p := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{}

		x.TTL--            //Decremeting the ttl for the packet
		x.DstIP = ip.DstIP //Changing encapsulated ip packet's destination address
		err := gopacket.SerializeLayers(p, opts, ip, gre, x)
		if err != nil {

			//handle error
		}
		packetsender <- p.Bytes() //sending packet to be send over the network
	}
}

//getBackendPool, returns backend pool as specified in configuration at intialization time
func getBackendPool() backendPool {

	return backendPool{}
}

var nextBackendIP chan net.IP

func nextBackEnd(ip net.IP, port layers.TCPPort) net.IP {

	return <-nextBackendIP
}

//Function to be used as go routine
func manageBackEnd() {
	nextBackendIP = make(chan net.IP)

	//Below code implements round robin technique
	for { //Infinte loop
		for _, v := range getBackendPool().pool {

			if v.isHealthy() {
				nextBackendIP <- v.IP
			}
		}
	}

}
