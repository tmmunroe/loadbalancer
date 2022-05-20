package loadbalancer

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

type Server struct {
	Name     string
	Address  *net.TCPAddr
	Listener net.Listener
	Server   *rpc.Server
	Running  bool
}

func InitServer(addr *net.TCPAddr) (*Server, error) {
	l, e := net.Listen(addr.Network(), addr.String())
	if e != nil {
		return nil, e
	}

	actualAddr, e := net.ResolveTCPAddr(l.Addr().Network(), l.Addr().String())
	if e != nil {
		return nil, e
	}

	return &Server{
		Address:  actualAddr,
		Listener: l,
		Server:   rpc.NewServer(),
	}, nil
}

func (s *Server) AddService(serv interface{}) error {
	return s.Server.Register(serv)
}

func (s *Server) Loop() error {
	step, _ := time.ParseDuration("5s")
	for s.Running {
		time.Sleep(step)
		log.Printf("Waiting...")
	}

	log.Printf("Stopped...")
	return nil
}

func (s *Server) Start() error {
	log.Printf("Accepting %v connections...", s.Address)
	go s.Server.Accept(s.Listener)
	return nil
}
