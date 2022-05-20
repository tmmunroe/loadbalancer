package main

import (
	"log"

	loadbalancer "github.com/tmmunroe/loadbalancer/pkg"
)

func main() {
	w, e := loadbalancer.InitWorker()
	if e != nil {
		log.Printf("error initializing worker: %v", e)
		return
	}
	w.Start()
}
