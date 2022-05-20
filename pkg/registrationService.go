package loadbalancer

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Registration struct {
	mu      sync.Mutex
	Servers []*net.TCPAddr
}

func InitRegistration() *Registration {
	return &Registration{
		mu:      sync.Mutex{},
		Servers: make([]*net.TCPAddr, 0),
	}
}

func (rs *Registration) Ping(args *PingArgs, reply *PingReply) error {
	log.Printf("Received Ping %v", args)
	return nil
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
