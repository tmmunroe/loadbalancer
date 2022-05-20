package loadbalancer

import "log"

type WorkerServices struct{}

func (w *WorkerServices) Ping(args *PingArgs, reply *PingReply) error {
	log.Printf("Received Ping %v", args)
	return nil
}

func (w *WorkerServices) Add(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Add %v", args)
	res := 0.0
	for _, x := range args.Numbers {
		res += x
	}
	reply.Answer = res
	log.Printf("Replying Add %v: %v", args, reply)
	return nil
}

func (w *WorkerServices) Multiply(args *MathArgs, reply *MathReply) error {
	log.Printf("Received Multiply %v", args)
	res := 1.0
	for _, x := range args.Numbers {
		res *= x
	}
	reply.Answer = res
	log.Printf("Replying Multiply %v: %v", args, reply)
	return nil
}
