package main

import (
	loadbalancer "github.com/tmmunroe/loadbalancer/pkg"
)

func main() {
	addr, regAddr := loadbalancer.LoadBalancerAddresses()
	lb := loadbalancer.InitLoadBalancer(addr, regAddr)

	lb.Start()
}
