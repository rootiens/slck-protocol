package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Printf("%v", err)
	}

	hub := newHub()
	go hub.run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}

		c := newClient(
			conn,
			hub.Commands,
			hub.Registrations,
			hub.Deregistrations,
		)

		go c.read()
	}
}
