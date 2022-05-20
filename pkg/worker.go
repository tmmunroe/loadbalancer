package loadbalancer

import (
	"log"
	"net"
	"net/rpc"
)

type Worker struct {
	Server *Server
}

func InitWorker() *Worker {
	a, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	s := InitServer(a)
	w := &Worker{Server: s}
	w.Server.AddService(w)
	return w
}

func (w *Worker) register() error {
	_, regAddr := LoadBalancerAddresses()
	c, e := rpc.Dial(regAddr.Network(), regAddr.String())
	if e != nil {
		log.Printf("error dialing register: %v", e)
		return e
	}

	log.Printf("registering")
	args := RegisterArgs{Address: regAddr}
	reply := RegisterReply{}
	e = c.Call("Registration.Register", &args, &reply)
	if e != nil {
		log.Printf("error registering: %v", e)
		return e
	}

	return nil
}

func (w *Worker) listen() error {
	return w.Server.Start()
}

func (w *Worker) Start() {
	e := w.register()
	if e != nil {
		log.Printf("error registering: %v", e)
		return
	}

	e = w.listen()
	if e != nil {
		log.Printf("error registering: %v", e)
		return
	}

	w.Server.Loop()
}

func (w *Worker) Ping(args *PingArgs, reply *PingReply) error {
	log.Printf("Received Ping %v", args)
	return nil
}

func (w *Worker) Add(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Add %v", args)
	res := 0.0
	for _, x := range args.Numbers {
		res += x
	}
	reply.Answer = res
	log.Printf("Replying Add %v: %v", args, reply)
	return nil
}

func (w *Worker) Multiply(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Multiply %v", args)
	res := 0.0
	for _, x := range args.Numbers {
		res *= x
	}
	reply.Answer = res
	log.Printf("Replying Multiply %v: %v", args, reply)
	return nil
}
