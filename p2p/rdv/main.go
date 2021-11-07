package main

import (
	"fmt"
	"log"

	"p2p/hole_punching/server"
)

func main() {
	fmt.Println("UDP Hole Punching Rendez-Vous Server")

	udpServer, err := server.NewServer("0.0.0.0:9001")
	if err != nil {
		log.Fatal(err)
	}

	udpServer.Listen()
}
