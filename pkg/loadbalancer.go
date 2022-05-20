package loadbalancer

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type ServiceRequest struct {
	Method  string
	Request interface{}
	Reply   interface{}
	Result  chan ServiceRequest
}

type MathServices struct {
	Requests chan ServiceRequest
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

	Done map[*net.TCPAddr]chan bool
}

func LoadBalancerAddresses() (*net.TCPAddr, *net.TCPAddr) {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55101")
	r, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:55103")
	return a, r
}

func InitMathService() *MathServices {
	return &MathServices{
		Requests: make(chan ServiceRequest, 100),
	}
}

func InitRegistration() *Registration {
	return &Registration{
		mu:      sync.Mutex{},
		Servers: make([]*net.TCPAddr, 0),
	}
}

func InitLoadBalancer(addr *net.TCPAddr, registrationAddr *net.TCPAddr) *LoadBalancer {
	s := InitServer(addr)
	m := InitMathService()
	s.AddService(m)

	rs := InitServer(registrationAddr)
	r := InitRegistration()
	rs.AddService(r)

	return &LoadBalancer{
		RegistrationServer: rs,
		Register:           r,
		ClientServer:       s,
		MathServices:       m,
		Done:               make(map[*net.TCPAddr]chan bool),
	}
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

func (l *LoadBalancer) startServerManager(addr *net.TCPAddr, done chan bool) error {
	log.Printf("starting server manager for %v", addr)
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("failed to connect to worker: %v", e)
		return e
	}

	for {
		log.Printf("waiting for requests to send to %v...", addr)
		select {
		case <-done:
			return nil
		case r := <-l.MathServices.Requests:
			log.Printf("Serving request %v on server %v", addr, r)
			e = c.Call(r.Method, r.Request, r.Reply)
			if e != nil {
				log.Printf("failed to call to worker: %v", e)
				return e
			}
			log.Printf("Returning request %v on server %v", addr, r)
			r.Result <- r
		}
	}
}

func (l *LoadBalancer) startServerManagers() {
	l.Register.mu.Lock()
	defer l.Register.mu.Unlock()
	for _, a := range l.Register.Servers {
		if _, e := l.Done[a]; e {
			continue
		}

		l.Done[a] = make(chan bool, 1)
		go l.startServerManager(a, l.Done[a])
	}
}

func (l *LoadBalancer) Loop() error {
	step, _ := time.ParseDuration("5s")
	for {
		time.Sleep(step)
		log.Printf("Requests: %v", len(l.MathServices.Requests))
		log.Printf("Registered Servers: %v", l.Register.Report())
		l.startServerManagers()
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
	sr := ServiceRequest{"Worker.Add", args, reply, result}

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
	sr := ServiceRequest{"Worker.Multiply", args, reply, result}

	log.Printf("Enqueueing Multiply %v", args)
	l.Requests <- sr

	log.Printf("Awaiting result Multiply %v", args)
	res := <-result
	rest := res.Reply.(*MathReply)

	log.Printf("Received result Multiply %v: %v", args, reply)
	reply.Answer = rest.Answer

	return nil
}
