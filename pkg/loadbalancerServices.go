package loadbalancer

import (
	"log"
	"sync"
	"time"
)

type LoadBalancerServices struct {
	mu       sync.Mutex
	Requests chan ServiceRequest
}

func InitLoadBalancerServices() *LoadBalancerServices {
	return &LoadBalancerServices{
		mu:       sync.Mutex{},
		Requests: make(chan ServiceRequest, 100),
	}
}

func (l *LoadBalancerServices) Add(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Add %v", args)
	result := make(chan ServiceRequest)
	timeout, _ := time.ParseDuration("1m")
	sr := ServiceRequest{"WorkerServices.Add", args, reply, result, timeout}

	log.Printf("Enqueueing Add %v", args)
	l.Requests <- sr

	log.Printf("Awaiting result Add %v", args)
	res := <-result
	rest := res.Reply.(*MathReply)

	log.Printf("Received result Add %v: %v", args, reply)
	reply.Answer = rest.Answer

	return nil
}

func (l *LoadBalancerServices) Multiply(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Multiply %v", args)
	result := make(chan ServiceRequest)
	timeout, _ := time.ParseDuration("1m")
	sr := ServiceRequest{"WorkerServices.Multiply", args, reply, result, timeout}

	log.Printf("Enqueueing Multiply %v", args)
	l.Requests <- sr

	log.Printf("Awaiting result Multiply %v", args)
	res := <-result
	rest := res.Reply.(*MathReply)

	log.Printf("Received result Multiply %v: %v", args, reply)
	reply.Answer = rest.Answer

	return nil
}
