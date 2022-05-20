package loadbalancer

import "net"

type PingArgs struct {
	From net.TCPAddr
}

type PingReply struct {
	From net.TCPAddr
}

type MathArgs struct {
	Numbers []float64
}

type MathReply struct {
	Answer float64
}

type RegisterArgs struct {
	Address *net.TCPAddr
}

type RegisterReply struct {
}
