package main

import (
 "bufio"
 "encoding/json"
 "fmt"
 "log"
 "net"
 "os"
 "time"
)

type Request struct {
 Method string        `json:"method"`
 Params []interface{} `json:"params"`
 ID     int           `json:"id"`
}

type Response struct {
 ID     int           `json:"id"`
 Result interface{}   `json:"result"`
 Error  interface{}   `json:"error"`
}

func main() {
 // Connect to the server
 conn, err := net.Dial("tcp", "localhost:12345")
 if err != nil {
 	log.Fatalf("Error connecting to server: %v", err)
 }
 defer conn.Close()

 // Create a buffered reader for the connection
 reader := bufio.NewReader(conn)

 // Subscribe to the server
 subscribeReq := Request{
 	Method: "mining.subscribe",
 	Params: []interface{}{},
 	ID:     1,
 }
 if err := json.NewEncoder(conn).Encode(subscribeReq); err != nil {
 	log.Fatalf("Error sending subscribe request: %v", err)
 }

 // Read the server response
 response, err := reader.ReadString('\n')
 if err != nil {
 	log.Fatalf("Error reading subscribe response: %v", err)
 }
 fmt.Printf("Subscribe response: %s\n", response)

 // Authorize the miner
 authorizeReq := Request{
 	Method: "mining.authorize",
 	Params: []interface{}{"miner1", "password"},
 	ID:     2,
 }
 if err := json.NewEncoder(conn).Encode(authorizeReq); err != nil {
 	log.Fatalf("Error sending authorize request: %v", err)
 }

 // Read the server response
 response, err = reader.ReadString('\n')
 if err != nil {
 	log.Fatalf("Error reading authorize response: %v", err)
 }
 fmt.Printf("Authorize response: %s\n", response)

 // Simulate job handling
 go func() {
 	for {
 		// Read job notifications from server
 		response, err := reader.ReadString('\n')
 		if err != nil {
 			log.Fatalf("Error reading job notification: %v", err)
 		}
 		fmt.Printf("Job notification: %s\n", response)

 		// Here you would normally process the job and submit a solution
 		// For this example, we just sleep for a while
 		time.Sleep(5 * time.Second)

 		// Submit a solution (dummy solution for this example)
 		submitReq := Request{
 			Method: "mining.submit",
 			Params: []interface{}{"miner1", "job-id", "nonce", "result"},
 			ID:     3,
 		}
 		if err := json.NewEncoder(conn).Encode(submitReq); err != nil {
 			log.Fatalf("Error sending submit request: %v", err)
 		}
 	}
 }()

 // Keep the client running
 select {}
