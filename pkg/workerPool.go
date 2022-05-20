package loadbalancer

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type WorkerHandle struct {
	Address            *net.TCPAddr
	Client             *rpc.Client
	Task               *ServiceRequest
	Call               *rpc.Call
	Requests           chan ServiceRequest
	Stop               chan bool
	Running            bool
	HealthCheckFreq    time.Duration
	HealthCheckTimeout time.Duration
}

type WorkerPool struct {
	mu      sync.Mutex
	Handles map[*net.TCPAddr]*WorkerHandle
}

func InitWorkerHandle(addr *net.TCPAddr, requests chan ServiceRequest, healthCheckFreq time.Duration, healthCheckTimeout time.Duration) *WorkerHandle {
	return &WorkerHandle{
		Address:            addr,
		Client:             nil,
		Task:               nil,
		Call:               nil,
		Requests:           requests,
		Stop:               make(chan bool, 1),
		Running:            false,
		HealthCheckTimeout: healthCheckTimeout,
		HealthCheckFreq:    healthCheckFreq,
	}
}

func InitWorkerPool() *WorkerPool {
	return &WorkerPool{
		mu:      sync.Mutex{},
		Handles: make(map[*net.TCPAddr]*WorkerHandle),
	}
}

func (wp *WorkerPool) AddWorkerHandle(addr *net.TCPAddr, requestFeed chan ServiceRequest, healthCheckFreq time.Duration, healthCheckTimeout time.Duration) (*WorkerHandle, error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if _, handleExists := wp.Handles[addr]; handleExists {
		return nil, fmt.Errorf("worker handle already exists for %v", addr.String())
	}

	handle := InitWorkerHandle(addr, requestFeed, healthCheckFreq, healthCheckTimeout)
	wp.Handles[addr] = handle

	return handle, nil
}

func (wh *WorkerHandle) stop() error {
	log.Printf("stopping worker handle %v", wh.Address.String())
	wh.Running = false
	if wh.Task != nil {
		log.Printf("re-enqueuing task")
		wh.Requests <- *wh.Task
		wh.Task = nil
		wh.Call = nil
	}

	if wh.Client != nil {
		return wh.Client.Close()
	}

	return nil
}

func (wh *WorkerHandle) serviceRequest(r ServiceRequest) error {
	log.Printf("Serving request %v on server %v", wh.Address.String(), r)
	wh.Task = &r
	call := wh.Client.Go(r.Method, r.Request, r.Reply, nil)
	select {
	case <-call.Done:
		log.Printf("Returning request %v on server %v", wh.Address.String(), r)
		r.Result <- r
		return call.Error

	case <-time.After(r.Timeout):
		return fmt.Errorf("call timed out")

	case <-wh.Stop:
		return wh.stop()
	}
}

func (wh *WorkerHandle) healthCheck() error {
	log.Printf("health check %v", wh.Address.String())
	p := &PingArgs{From: *wh.Address}
	r := &PingReply{}
	call := wh.Client.Go("WorkerServices.Ping", p, r, nil)
	select {
	case <-call.Done:
		log.Printf("ping complete %v", wh.Address.String())
		return call.Error

	case <-time.After(wh.HealthCheckTimeout):
		return fmt.Errorf("ping timed out %v", wh.Address.String())
	}
}

func (wh *WorkerHandle) Run() {
	log.Printf("starting handle for %v", wh.Address.String())
	c, e := rpc.Dial(wh.Address.Network(), wh.Address.String())
	if e != nil {
		log.Fatalf("error dialing: %v", e)
	}

	wh.Running = true
	wh.Client = c

	for wh.Running {
		log.Printf("waiting for requests to send to %v...", wh.Address.String())

		select {
		case <-wh.Stop:
			wh.stop()

		case <-time.After(wh.HealthCheckFreq):
			hce := wh.healthCheck()
			if hce != nil {
				log.Printf("stopping worker handle %v due to failed health check: %v", wh.Address.String(), hce)
				wh.stop()
			}

		case r := <-wh.Requests:
			e = wh.serviceRequest(r)
			if e != nil {
				log.Printf("service request failed: %v... requeing request %v", e, r)
				wh.Requests <- r
			}
			wh.Task = nil
		}

	}
}
