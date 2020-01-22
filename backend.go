package main

import (
	"sync"

	"github.com/devansh42/sm"
	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

//initBackend, initiates backend node to handle load
func initBackend() {

	ss := sm.NewSequentialServiceManager()
	ss.AddService(sm.Service{Executer: packetSenderListner})
	ss.AddService(sm.Service{Executer: handleBackendIngressTraffic})
	ss.AddService(sm.Service{Executer: func() {
		//For logging purposes
		glog.Info("Backend is Ready at ", props.port)
	}})
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		err := ss.Start()
		if err != nil {
			glog.Fatal("Couldn't initalize backend :", err)
			wg.Done()
		}

	}()
	wg.Wait()
}

func handleBackendIngressTraffic() {
	glog.Info("Ready to accept Incoming Packets")
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

//globalbackendPool, is channel to exchange backend pool instance
//so that we can dynamically append remove backend instances
var globalbackendPool *backendPool
