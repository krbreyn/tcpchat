package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// todo - grab and print local ip on startup so you know what to connect to. also say 'localhost' for clarity

func main() {
	// server vars
	numClients := 0
	clients := make(map[net.Conn]int) // int is client id
	newClients := make(chan net.Conn)
	deadClients := make(chan net.Conn)
	msgs := make(chan string)

	port := ":23"

	server, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("running on %s", port)

	// accept connections
	go func() {
		for {
			// TODO just handle and print to console
			conn, err := server.Accept()
			if err != nil {
				panic(err)
			}
			newClients <- conn
		}
	}()

	// main loop
	for {
		select {
		// new clients
		case conn := <-newClients:
			log.Printf("accepted client %d", numClients)
			clients[conn] = numClients
			numClients++

			// read messages from client
			go func(conn net.Conn, id int) {
				reader := bufio.NewReader(conn)
				for {
					msg, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					msgs <- fmt.Sprintf("client %d > %s", id, msg)
				}
				// if broken, client is dead
				deadClients <- conn
			}(conn, clients[conn])

		// dead clients
		case conn := <-deadClients:
			log.Printf("disconnected client %v", clients[conn])
			delete(clients, conn)

		// handle messages
		case msg := <-msgs:
			for conn := range clients {
				go func(conn net.Conn, msg string) {
					_, err := conn.Write([]byte(msg))
					if err != nil {
						deadClients <- conn
					}
				}(conn, msg)
			}
			log.Printf("msg: %s", msg)
			log.Printf("%d clients connected", len(clients))
		}
	}
}
