package loadbalancer

import (
	"log"
	"net/rpc"
)

type Client struct {
}

func (c *Client) mathService(serviceMethod string, numbers ...float64) (float64, error) {
	log.Printf("Math service %v...", serviceMethod)

	addr, _ := LoadBalancerAddresses()
	log.Printf("Dialing load balancer %v %v...", addr.Network(), addr.String())
	d, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("error dialing: %v", e)
		return -1, e
	}
	defer d.Close()

	log.Printf("Calling math service %v...", serviceMethod)
	args := &MathArgs{}
	reply := &MathReply{}
	e = d.Call(serviceMethod, args, reply)
	if e != nil {
		log.Printf("error calling: %v", e)
		return -1, e
	}

	log.Printf("Returning reply %v...", serviceMethod)
	return reply.Answer, nil
}

func (c *Client) Add(numbers ...float64) (float64, error) {
	return c.mathService("MathServices.Add", numbers...)
}

func (c *Client) Multiply(numbers ...float64) (float64, error) {
	return c.mathService("MathServices.Multiply", numbers...)
}
