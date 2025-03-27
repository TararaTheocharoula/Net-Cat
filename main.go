package main

import (
	"fmt"
	"log"
	"net-cat/server"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	srv := server.NewServer(port)
	defer srv.Close()

	if err := srv.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}
