package main

import (
	"fmt"

	loadbalancer "github.com/tmmunroe/loadbalancer/pkg"
)

func main() {
	client := loadbalancer.Client{}
	fmt.Print(client.Add(10, 1))
}
