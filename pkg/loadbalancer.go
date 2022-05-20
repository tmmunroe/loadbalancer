package loadbalancer

import (
	"log"
	"net"
	"time"
)

type ServiceRequest struct {
	Method  string
	Request interface{}
	Reply   interface{}
	Result  chan ServiceRequest
	Timeout time.Duration
}

type LoadBalancer struct {
	RegistrationServer *Server
	Register           *Registration

	ClientServer         *Server
	LoadBalancerServices *LoadBalancerServices

	WorkerPool *WorkerPool

	Running bool
}

func LoadBalancerAddresses() (*net.TCPAddr, *net.TCPAddr) {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55101")
	r, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55103")
	return a, r
}

func InitLoadBalancer(addr *net.TCPAddr, registrationAddr *net.TCPAddr) (*LoadBalancer, error) {
	s, e := InitServer(addr)
	if e != nil {
		return nil, e
	}

	rs, e := InitServer(registrationAddr)
	if e != nil {
		return nil, e
	}

	m := InitLoadBalancerServices()
	s.AddService(m)

	r := InitRegistration()
	rs.AddService(r)

	wp := InitWorkerPool()

	return &LoadBalancer{
		RegistrationServer:   rs,
		Register:             r,
		ClientServer:         s,
		LoadBalancerServices: m,
		WorkerPool:           wp,
	}, nil
}

func (l *LoadBalancer) startWorkerHandles() {
	l.Register.mu.Lock()
	defer l.Register.mu.Unlock()

	for _, a := range l.Register.Servers {
		if _, e := l.WorkerPool.Handles[a]; e {
			continue
		}

		frequency, _ := time.ParseDuration("10s")
		timeout, _ := time.ParseDuration("10s")
		handle, e := l.WorkerPool.AddWorkerHandle(a, l.LoadBalancerServices.Requests, frequency, timeout)
		if e != nil {
			log.Panicf("could not add worker handle %v", a)
		}

		go handle.Run()
	}
}

func (l *LoadBalancer) Loop() error {
	step, _ := time.ParseDuration("5s")
	for l.Running {
		time.Sleep(step)
		log.Printf("Requests: %v", len(l.LoadBalancerServices.Requests))
		log.Printf("Registered Servers: %v", l.Register.Report())
		l.startWorkerHandles()
	}

	return nil
}

func (l *LoadBalancer) Start() error {
	log.Printf("Starting load balancer")
	l.Running = true
	l.ClientServer.Start()
	l.RegistrationServer.Start()
	return l.Loop()
}
