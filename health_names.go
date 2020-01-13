package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const ipKey uint8 = 0

//noOfChecks, is the no for checks for one health check
const noOfChecks = 10

//instanceHealthChecker, is health checker for instance whether it is backend or load balancer
type instanceHealthChecker interface {
	performHealthCheck()

	//healthCheckPerformer, Returns function which is used for health checking for each instance
	healthCheckPerformer() func(context.Context, backend) <-chan error

	//getTTI, returns Timeout,Thresold,Interval,
	getTTI() (time.Duration, float32, time.Duration)
}

//basicHealthChecker contains to attribute to hold about a health check
type basicHealthChecker struct {
	port      uint
	timeout   time.Duration
	interval  time.Duration
	threshold float32
}

//getTTI, returns Timeout,Thresold and Interval
func (b basicHealthChecker) getTTI() (time.Duration, float32, time.Duration) {
	return b.timeout, b.threshold, b.interval
}

//generalHealthChecker, is the general health checker for instances
type generalHealthChecker struct {
	basicHealthChecker
}

func (g generalHealthChecker) isHealthy() bool {

	return true
}

//performHealthCheck, perfroms health of
func (g generalHealthChecker) performHealthCheck() {
}

//tcphealthchecker, is the health checker via tcp
type tcphealthchecker struct {
	tcporudphealthcheck
}

//httphealthchecker, is the health checker via http
type httphealthchecker struct {
	basicHealthChecker
	//path, is the path of http endpoint
	path string
	//method, is the http method for request
	method string
	//niceStatus, the status code which states about healthy response
	niceStatus int
}

//udphealthchecker, is the health checker via udp
type udphealthchecker struct {
	tcporudphealthcheck
}

//tcporudphealthcheck, is the health checker for tcp or udp protocol
type tcporudphealthcheck struct {
	basicHealthChecker
	//protocol, specifies the protocol for health check 0 for tcp and 1 for udp
	protocol uint
}

//Errors
var (
	timeOutErr                 = errors.New("Request Timeout")
	couldntEstablishConnection = errors.New("Couldn't establish connection")
	notHealthyStatus           = errors.New("Not a healthy Status")
)

func (t tcporudphealthcheck) healthCheckPerformer() func(context.Context, backend) <-chan error {
	return func(con context.Context, b backend) <-chan error {
		c := make(chan error)
		go func() {
			var pp string
			if t.protocol == 1 {
				pp = "udp"
			} else {
				pp = "tcp"
			}

			conn, err := net.Dial(pp, fmt.Sprint(b.IP.String(), b.Port))
			if err != nil {
				c <- err
				return
			}

			defer conn.Close() //Closing the connection
			var bb []byte = []byte("PING")
			_, err = conn.Write(bb)
			if err != nil {
				c <- err
				return
			}
			c <- nil
		}()
		return c
	}
}

func (t tcporudphealthcheck) performHealthCheck() {

}

func (t httphealthchecker) healthCheckPerformer() func(context.Context, backend) <-chan error {
	return func(con context.Context, b backend) <-chan error {
		c := make(chan error)
		go func() {
			r := new(http.Request)
			if t.method == "" {
				t.method = "get"
			}
			r.Method = t.method
			r = r.WithContext(con)
			r.URL, _ = url.Parse(fmt.Sprint("http://", b.IP.String(), ":", b.Port, t.path))
			cli := http.DefaultClient
			//set timeout for request
			resp, err := cli.Do(r)
			if err != nil {
				c <- err
				return
				//handle error
			}
			if resp.StatusCode != t.niceStatus {
				//handle error
				c <- notHealthyStatus
				return
			}
			c <- nil //No error
		}()
		return c
	}

}

func (t httphealthchecker) performHealthCheck() {
}
