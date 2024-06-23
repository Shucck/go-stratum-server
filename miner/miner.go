package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)


type Miner struct {
	ID       string
	password string
	conn     net.Conn
}

func NewMiner(id, password string) *Miner {
	return &Miner{
		ID:       id,
		password: password,
	}
}

//connects the miner to the server.
func (m *Miner) Connect(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	m.conn = conn
	return nil
}

func (m *Miner) Subscribe() error {
	request := map[string]interface{}{
		"id":     1,
		"method": "mining.subscribe",
		"params": []interface{}{},
	}
	return m.sendRequest(request)
}

func (m *Miner) Authorize() error {
	request := map[string]interface{}{
		"id":     2,
		"method": "mining.authorize",
		"params": []interface{}{m.ID, m.password},
	}
	return m.sendRequest(request)
}

// Submit solution
func (m *Miner) Submit(jobID, nonce, result string) error {
	request := map[string]interface{}{
		"id":     3,
		"method": "mining.submit",
		"params": []interface{}{m.ID, jobID, nonce, result},
	}
	return m.sendRequest(request)
}

//JSON-RPC request
func (m *Miner) sendRequest(request map[string]interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_, err = m.conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (m *Miner) receiveResponse() {
	for {
		buffer := make([]byte, 1024)
		n, err = m.conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			return
		}
		var response map[string]interface{}
		if err := json.Unmarshal(buffer[:n], &response); err != nil {
			log.Printf("Error decoding response: %v", err)
			continue
		}
		log.Printf("Received response: %v", response)
	}
}

func main() {
	miner := NewMiner("Shuccck", "stratumServer")
	err := miner.Connect("localhost", 3333)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}

	err = miner.Subscribe()
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	err = miner.Authorize()
	if err != nil {
		log.Fatalf("Error authorizing: %v", err)
	}

	go miner.receiveResponse()

	time.Sleep(10 * time.Second)
	err = miner.Submit("job-1", "00000001", "solution")
	if err != nil {
		log.Fatalf("Error submitting solution: %v", err)
	}
}
