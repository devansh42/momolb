package main

import (
	"context"
	"sync"
	"time"
)

//reportCardGenerator, generates a report card for a healthcheck session
func reportCardGenerator(accumaltor <-chan report, th float32, pool *backendPool) {
	var reportCard = make(map[string]uint8)
	for x := range accumaltor {
		if x.err != nil {
			if _, ok := reportCard[x.backend.Name]; ok {
				reportCard[x.backend.Name]++
			} else {
				reportCard[x.backend.Name] = 1
			}
		}
	}
	for k, v := range reportCard {
		var b backend
		pool.RLock()
		for _, x := range pool.pool {
			if x.Name == k {
				b = x
				break
			}
		}
		pool.RUnlock()
		if (float32(v) / float32(noOfChecks)) >= th {
			//In HealthyState
			pool.markHealthy(b)
		} else {
			pool.markUnHealthy(b)
		}

	}

}

type report struct {
	err     error
	backend backend
}

//healthChecker, is the health checker which runs in bakeground and checks health periodically
func healthChecker(pool *backendPool) {

	pool.RLock()
	//Read only part
	checker := pool.healthChecker
	performer := pool.healthChecker.healthCheckPerformer()
	poolItems := pool.pool

	pool.RUnlock()
	tim, th, _ := checker.getTTI()
	con := context.Background()

	accumaltor := make(chan report)
	go reportCardGenerator(accumaltor, th, pool)

	var wg = new(sync.WaitGroup)
	for _, b := range poolItems {

		for i := 0; i < noOfChecks; i++ {
			c, _ := context.WithTimeout(con, tim)
			ch := performer(c, b)
			wg.Add(1)
			go func(b backend) {

				select {
				case <-c.Done(): //It only closes when there is a Request timeout
					accumaltor <- report{context.DeadlineExceeded, b}
					wg.Done()

					break
				case err := <-ch: //It
					accumaltor <- report{err, b}
					wg.Done()

					break
				}
			}(b)
		}
	}
	wg.Wait()         //Waiting for results to be
	close(accumaltor) //As all the results are calculated it is the time to close the accumalator channel

}

//healthCheckService, performs health checks in background
func healthCheckService() {

	pool := <-globalbackendPool
	for {

		_, _, i := pool.healthChecker.getTTI()
		go healthChecker(pool)
		time.Sleep(i) //Wait for i duration

	}
}

//globalHealthChecker, is the global health checker for instance
var globalHealthCheckerCh = make(chan instanceHealthChecker)
