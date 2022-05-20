package main

import loadbalancer "github.com/tmmunroe/loadbalancer/pkg"

func main() {
	w := loadbalancer.InitWorker()
	w.Start()
}
