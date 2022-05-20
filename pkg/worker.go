package loadbalancer

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

type Worker struct {
	Server    *Server
	RegClient *rpc.Client
}

func InitWorker() (*Worker, error) {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	_, regAddr := LoadBalancerAddresses()

	s, e := InitServer(a)
	if e != nil {
		return nil, fmt.Errorf("error initializing worker services: %v", e)
	}

	c, e := rpc.Dial(regAddr.Network(), regAddr.String())
	if e != nil {
		return nil, fmt.Errorf("error dialing register: %v", e)
	}

	w := &Worker{
		Server:    s,
		RegClient: c,
	}

	ws := &WorkerServices{}
	e = w.Server.AddService(ws)
	if e != nil {
		w.RegClient.Close()
		w.RegClient = nil
		return nil, fmt.Errorf("error registering worker services: %v", e)
	}

	return w, nil
}

func (w *Worker) register() error {
	log.Printf("registering")
	args := RegisterArgs{Address: w.Server.Address}
	reply := RegisterReply{}

	e := w.RegClient.Call("Registration.Register", &args, &reply)
	if e != nil {
		return fmt.Errorf("error registering: %v", e)
	}

	return nil
}

func (w *Worker) listen() error {
	return w.Server.Start()
}

func (w *Worker) healthCheck() error {
	log.Printf("pinging load balancer...")
	args := PingArgs{From: *w.Server.Address}
	reply := PingReply{}

	e := w.RegClient.Call("Registration.Ping", &args, &reply)
	if e != nil {
		return fmt.Errorf("error registering: %v", e)
	}

	log.Printf("ping complete...")
	return nil
}

func (w *Worker) Loop() error {
	w.Server.Running = true
	step, _ := time.ParseDuration("5s")
	for w.Server.Running {
		time.Sleep(step)
		log.Printf("Waiting...")
		e := w.healthCheck()
		if e != nil {
			log.Printf("health check failed: %v", e)
			w.Server.Running = false
			return e
		}
	}

	log.Printf("Stopped...")
	return nil
}

func (w *Worker) Start() {
	e := w.register()
	if e != nil {
		log.Printf("error registering: %v", e)
		return
	}

	e = w.listen()
	if e != nil {
		log.Printf("error listening: %v", e)
		return
	}

	log.Printf("looping indefinetly...")
	w.Loop()
}
