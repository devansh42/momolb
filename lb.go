package main

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	IPV4 = "ipv4"
)

var lbincomingPacket = make(chan *layers.IPv4)

//initLB, initiates Load Balancer
func initLB() {

	go handleLBIngress()
	go manageBackEnd()
	go packetSenderListner()

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
	return make(backendPool, 0)
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
		for _, v := range getBackendPool() {
			if v.HealthChecker.isHealthy() {
				nextBackendIP <- v.IP
			}
		}
	}

}
