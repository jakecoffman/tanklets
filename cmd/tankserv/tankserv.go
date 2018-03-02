package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jakecoffman/tanklets/server"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// TODO make port configurable
	port := "8999"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	network := server.NewServer("0.0.0.0:"+port)

	go network.Recv()
	defer func() { fmt.Println(network.Close()) }()

	fmt.Println("Server Running on", port)

	server.Lobby(network)
}
