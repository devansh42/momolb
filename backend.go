package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"stathat.com/c/consistent"
	"github.com/devansh42/sm"
)

type backend struct {
	Name          string
	IP            net.IP
	Port          uint16
	HealthChecker instanceHealthChecker
	healthy       bool
}

var backendincomingPacket = make(chan *layers.GRE)

//backendPool, is the pool for backend
type backendPool struct {
	pool          []backend
	healthChecker instanceHealthChecker
	consistency   *consistent.Consistent
}

//onlyHealty, returns only healthy backends
func (p backendPool) onlyHealty() backendPool {
	var h = make([]backend, len(p.pool))
	var count = 0
	for _, v := range p.pool {
		if v.health() {
			h[count] = v
			count++
		}
	}
	return backendPool{pool: h[:count]}
}

//add, adds new backend instance
//healthchecking method is same for instances in a backend pool
//instance is in unhealthy state untill its first health check success
func (p *backendPool) add(b backend) {
	b.HealthChecker = p.healthChecker
	p.pool = append(p.pool, b)

}

//remove, removes backend from the backend pool
//It uses backend ip for matching
//Returns removed backend
func (p *backendPool) remove(b backend) backend {
	var pool = make([]backend, len(p.pool))
	var r backend
	count := 0
	for _, v := range p.pool {
		if v.Name == b.Name {
			r = v
			p.consistency.Remove(b.Name)
			continue
		} else {
			pool[count] = v
			count++
		}
	}
	p.pool = pool[:count]
	return r
}

//markHealthy, marks a backend instance healthy so it listen traffic
func (p backendPool) markHealthy(b backend) {
	for _, v := range p.pool {
		if v.Name == b.Name && !v.healthy {
			v.healthy = true
			p.consistency.Add(b.Name) //Removing from load balancer
			break
		}
	}
}

//markUnHealthy, marks a backend instance unhealthy lb stops forwarding load on it
func (p backendPool) markUnHealthy(b backend) {
	for _, v := range p.pool {
		if v.Name == b.Name && v.healthy {
			v.healthy = false
			p.consistency.Remove(b.Name) //Adding to listen load balancer
			break
		}
	}
}


















//initBackend, initiates backend node to handle load
func initBackend() {
	var x = sm.NewServiceManager()
	x.AddService(handleBackendIngressTraffic)
	x.AddService(packetSenderListner)
	x.Start()
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
