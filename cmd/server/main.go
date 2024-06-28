package main

import (
	"log"
	"time"

	"github.com/Shucck/go-stratum-server/main/server"
)

func main() {
	host := "localhost"
	port := 8080
	network := "tcp"
	connectionTimeout := 5 * time.Minute
	maxWorkers := 10

	stratumServer := server.NewStratumServer(host, port, network, connectionTimeout, maxWorkers)
	go stratumServer.Start()

	log.Printf("Stratum server running on %s:%d\n", host, port)

	// Keep the main goroutine running
	select {}
}
