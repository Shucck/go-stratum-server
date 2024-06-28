package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func main() {
	// Connect to the Stratum server
	conn, err := net.Dial("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to Stratum server")

	// Step 1: Send mining.subscribe request
	subscribeMessage := map[string]interface{}{
		"id":     1,
		"method": "mining.subscribe",
		"params": []interface{}{},
	}
	if err := sendJSON(conn, subscribeMessage); err != nil {
		log.Fatalf("Error sending subscribe message: %v", err)
	}

	// Step 2: Wait for subscription response
	response, err := receiveJSON(conn)
	if err != nil {
		log.Fatalf("Error receiving subscribe response: %v", err)
	}
	log.Printf("Subscribe response received: %v", response)

	// Extract subscription ID and extranonce1
	subscriptionID := response["result"].([]interface{})[0].(string)
	extranonce1 := response["result"].([]interface{})[1].(string)

	// Step 3: Send mining.authorize request with credentials
	authorizeMessage := map[string]interface{}{
		"id":     2,
		"method": "mining.authorize",
		"params": []interface{}{"Shuccck", "stratumServer"},
	}
	if err := sendJSON(conn, authorizeMessage); err != nil {
		log.Fatalf("Error sending authorize message: %v", err)
	}

	// Step 4: Wait for authorization response
	response, err = receiveJSON(conn)
	if err != nil {
		log.Fatalf("Error receiving authorize response: %v", err)
	}
	log.Printf("Authorize response received: %v", response)

	// Step 5: Wait for mining.notify messages with new jobs
	for {
		notifyMessage, err := receiveJSON(conn)
		if err != nil {
			log.Fatalf("Error receiving notify message: %v", err)
		}
		if notifyMessage["method"].(string) == "mining.notify" {
			log.Printf("Received mining job: %v", notifyMessage)
			// Step 6: Start mining using job parameters
			// Step 7: Once valid nonce is found, send mining.submit request with solution
		}
	}
}

// Helper function to send JSON-RPC messages
func sendJSON(conn net.Conn, message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	_, err = fmt.Fprintf(conn, "%s\n", data)
	if err != nil {
		return fmt.Errorf("error sending JSON message: %v", err)
	}
	return nil
}

// Helper function to receive JSON-RPC messages
func receiveJSON(conn net.Conn) (map[string]interface{}, error) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var message map[string]interface{}
		err := json.Unmarshal(scanner.Bytes(), &message)
		if err != nil {
			return nil, fmt.Errorf("error decoding JSON: %v", err)
		}
		return message, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from connection: %v", err)
	}
	return nil, nil
}
