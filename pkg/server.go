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
}

func InitServer(addr *net.TCPAddr) *Server {
	return &Server{
		Address: addr,
		Server:  rpc.NewServer(),
	}
}

func (s *Server) AddService(serv interface{}) error {
	return s.Server.Register(serv)
}

func (s *Server) Loop() error {
	step, _ := time.ParseDuration("5s")
	for {
		time.Sleep(step)
		log.Printf("Waiting...")
	}
}

func (s *Server) Start() error {
	log.Printf("Starting %v server %v %v...", s.Name, s.Address.Network(), s.Address.String())
	l, e := net.Listen(s.Address.Network(), s.Address.String())
	if e != nil {
		return e
	}

	s.Listener = l
	log.Printf("Accepting %v connections...", s.Name)
	go s.Server.Accept(l)

	return nil
}
