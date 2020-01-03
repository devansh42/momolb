package main

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const KB = 1024

type packetdata []byte

var packetsender = make(chan packetdata)

//packetSenderListner, this simply sends packet to the destination
func packetSenderListner() {

	conn, _ := net.DialIP(IPV4, nil, nil)
	for x := range packetsender {
		_, err := conn.Write(x)
		if err != nil {
			//handle error
		}
	}

}

//layercontent, contains content and payload for TCP Packet
//Captured by pcap
func initTrafficInspetion() {

	//Reading live packets from eth0 interface
	//here eth0 is the default interface
	handler, err := pcap.OpenLive("eth0", KB, false, pcap.BlockForever)
	if err != nil {
		panic(err)
		return
	}
	var filter string
	if props.islb {
		filter = fmt.Sprint("tcp dst port ", props.port)
	} else {
		//filter for gre
		//47 is the protocol number for GRE
		//for more protocols number see at /etc/protocols
		filter = fmt.Sprint("ip proto 47")
	}
	err = handler.SetBPFFilter(filter)
	src := gopacket.NewPacketSource(handler, handler.LinkType())
	for packet := range src.Packets() {
		if props.islb {
			handlePacket(packet)

		} else {
			handleBackendPacket(packet)
		}

	}
}

func handleBackendPacket(packet gopacket.Packet) {
	grelayer := packet.Layer(layers.LayerTypeGRE)
	if grelayer != nil {
		if gre, ok := grelayer.(*layers.GRE); ok {
			//There is gre layers
			backendincomingPacket <- gre
		}
	}
}

//handlePacket, handles packet handling
func handlePacket(packet gopacket.Packet) {
	iplayer := packet.Layer(layers.LayerTypeIPv4)
	if iplayer != nil {

		if ippack, ok := iplayer.(*layers.IPv4); ok {
			//There is an IP layer
			//Now send this packet data to its destination

			lbincomingPacket <- ippack
		}
	}
}
