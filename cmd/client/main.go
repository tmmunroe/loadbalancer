package main

import (
	"fmt"
	"log"
	"sync"

	loadbalancer "github.com/tmmunroe/loadbalancer/pkg"
)

func oneClient() {
	c, e := loadbalancer.InitClient()
	if e != nil {
		log.Printf("error initializing client: %v", e)
	}
	r, e := c.Add(10, 1)
	fmt.Printf("Replied: %v, %v", r, e)
}

func manyClients() {
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	clientLimit := 10
	requestsPerClient := 100
	clients := make([]*loadbalancer.Client, 0)
	for i := 0; i < clientLimit; i++ {
		c, e := loadbalancer.InitClient()
		if e != nil {
			log.Printf("error initializing client %v: %v", i, e)
		}
		clients = append(clients, c)
	}

	for j := 1; j < requestsPerClient; j++ {
		for i, client := range clients {
			wg.Add(2)
			go func(i int, c *loadbalancer.Client) {
				mu.Lock()
				log.Printf("submitting task %v client %v", j, i)
				mu.Unlock()
				r, e := c.Add(10, 1)
				mu.Lock()
				log.Printf("Add Replied: %v, %v", r, e)
				mu.Unlock()
				wg.Done()
			}(i, client)

			go func(i int, c *loadbalancer.Client) {
				mu.Lock()
				log.Printf("submitting task %v client %v", j, i)
				mu.Unlock()
				r, e := c.Multiply(10, 1)
				mu.Lock()
				log.Printf("Multiply Replied: %v, %v", r, e)
				mu.Unlock()
				wg.Done()
			}(i, client)
		}
	}
	wg.Wait()
}

func main() {
	manyClients()
}
