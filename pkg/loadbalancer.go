package loadbalancer

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ServiceRequest struct {
	Method  string
	Request interface{}
	Reply   interface{}
	Result  chan ServiceRequest
	Timeout time.Duration
}

type MathServices struct {
	Requests chan ServiceRequest

	mu      sync.Mutex
	Pending map[*net.TCPAddr]*ServiceRequest
}

type Registration struct {
	mu      sync.Mutex
	Servers []*net.TCPAddr
}

type LoadBalancer struct {
	RegistrationServer *Server
	Register           *Registration

	ClientServer *Server
	MathServices *MathServices

	WorkerPool *WorkerPool
}

func LoadBalancerAddresses() (*net.TCPAddr, *net.TCPAddr) {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55101")
	r, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55103")
	return a, r
}

func InitMathService() *MathServices {
	return &MathServices{
		Requests: make(chan ServiceRequest, 100),
		mu:       sync.Mutex{},
		Pending:  make(map[*net.TCPAddr]*ServiceRequest),
	}
}

func InitRegistration() *Registration {
	return &Registration{
		mu:      sync.Mutex{},
		Servers: make([]*net.TCPAddr, 0),
	}
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

	m := InitMathService()
	s.AddService(m)

	r := InitRegistration()
	rs.AddService(r)

	wp := InitWorkerPool()

	return &LoadBalancer{
		RegistrationServer: rs,
		Register:           r,
		ClientServer:       s,
		MathServices:       m,
		WorkerPool:         wp,
	}, nil
}

func (rs *Registration) Register(args *RegisterArgs, reply *RegisterReply) error {
	log.Printf("Registering %v %v...", args.Address.Network(), args.Address.String())
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.Servers = append(rs.Servers, args.Address)
	log.Printf("Registered %v %v...", args.Address.Network(), args.Address.String())
	return nil
}

func (rs *Registration) Report() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return fmt.Sprintf("Registration: %v", rs.Servers)
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
		handle, e := l.WorkerPool.AddWorkerHandle(a, l.MathServices.Requests, frequency, timeout)
		if e != nil {
			log.Panicf("could not add worker handle %v", a)
		}

		go handle.Run()
	}
}

func (l *LoadBalancer) Loop() error {
	step, _ := time.ParseDuration("5s")
	for {
		time.Sleep(step)
		log.Printf("Requests: %v", len(l.MathServices.Requests))
		log.Printf("Registered Servers: %v", l.Register.Report())
		l.startWorkerHandles()
	}
}

func (l *LoadBalancer) Start() error {
	log.Printf("Starting load balancer")
	l.ClientServer.Start()
	l.RegistrationServer.Start()
	return l.Loop()
}

func (l *LoadBalancer) Ping(args *PingArgs, reply *PingReply) error {
	log.Printf("Received Ping %v", args)
	return nil
}

func (l *MathServices) Add(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Add %v", args)
	result := make(chan ServiceRequest)
	timeout, _ := time.ParseDuration("1m")
	sr := ServiceRequest{"Worker.Add", args, reply, result, timeout}

	log.Printf("Enqueueing Add %v", args)
	l.Requests <- sr

	log.Printf("Awaiting result Add %v", args)
	res := <-result
	rest := res.Reply.(*MathReply)

	log.Printf("Received result Add %v: %v", args, reply)
	reply.Answer = rest.Answer

	return nil
}

func (l *MathServices) Multiply(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Multiply %v", args)
	result := make(chan ServiceRequest)
	timeout, _ := time.ParseDuration("1m")
	sr := ServiceRequest{"Worker.Multiply", args, reply, result, timeout}

	log.Printf("Enqueueing Multiply %v", args)
	l.Requests <- sr

	log.Printf("Awaiting result Multiply %v", args)
	res := <-result
	rest := res.Reply.(*MathReply)

	log.Printf("Received result Multiply %v: %v", args, reply)
	reply.Answer = rest.Answer

	return nil
}
