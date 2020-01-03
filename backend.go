package main

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type backend struct {
	Name          string
	IP            net.IP
	Port          uint16
	HealthChecker instanceHealthChecker
}

var consistencyCircle *

var backendincomingPacket = make(chan *layers.GRE)

//backendPool, is the pool for backend
type backendPool []backend

//initBackend, initiates backend node to handle load
func initBackend() {
	go handleBackendIngressTraffic()
	go packetSenderListner()
}

func handleBackendIngressTraffic() {
	for x := range backendincomingPacket {
		//This is an GRE Packet which contains encapsulated ip packet delivered to backend

		packet := gopacket.NewPacket(x.LayerPayload(), layers.LayerTypeIPv4, gopacket.Default)
		if packet != nil {
			//This means we have genuine gre encapsulated ip packet
			//Let's forward this packet

			packetsender <- packet.Data()
			//Above action sends encapsulated ip packet over the ip network of backend on
			//The tcp packet's dest port is same as lb listening port
		}
	}
}
