package main

import (
	"log"

	loadbalancer "github.com/tmmunroe/loadbalancer/pkg"
)

func main() {
	addr, regAddr := loadbalancer.LoadBalancerAddresses()
	lb, e := loadbalancer.InitLoadBalancer(addr, regAddr)
	if e != nil {
		log.Panicf("error initializing load balancer: %v", e)
	}

	lb.Start()
}
