package loadbalancer

import (
	"log"
	"net/rpc"
)

type Client struct {
	client *rpc.Client
}

func InitClient() (*Client, error) {
	addr, _ := LoadBalancerAddresses()
	log.Printf("Dialing load balancer %v %v...", addr.Network(), addr.String())
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("error dialing: %v", e)
		return nil, e
	}

	return &Client{client: c}, nil
}

func (c *Client) mathService(serviceMethod string, numbers ...float64) (float64, error) {

	//log.Printf("Math service %v...", serviceMethod)

	//log.Printf("Calling math service %v...", serviceMethod)
	args := &MathArgs{Numbers: numbers}
	reply := &MathReply{}
	e := c.client.Call(serviceMethod, args, reply)
	if e != nil {
		log.Printf("error calling: %v", e)
		return -1, e
	}

	//log.Printf("Returning reply %v...", serviceMethod)
	return reply.Answer, nil
}

func (c *Client) Add(numbers ...float64) (float64, error) {
	return c.mathService("LoadBalancerServices.Add", numbers...)
}

func (c *Client) Multiply(numbers ...float64) (float64, error) {
	return c.mathService("LoadBalancerServices.Multiply", numbers...)
}
