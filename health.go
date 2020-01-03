package main

//instanceHealthChecker, is health checker for instance whether it is backend or load balancer
type instanceHealthChecker interface {
	isHealthy() bool
}

//generalHealthChecker, is the general health checker for instances
type generalHealthChecker struct{}

func (g generalHealthChecker) isHealthy() bool {
	return true
}
