package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type Job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type Solution struct {
	MinerID   string `json:"minerId"`
	JobID     string `json:"jobId"`
	Nonce     string `json:"nonce"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

type StratumServer struct {
	host      string
	port      int
	miners    map[string]net.Conn
	minerPass map[string]string
	mutex     sync.Mutex
	jobID     uint64
}

func NewStratumServer(host string, port int) *StratumServer {
	return &StratumServer{
		host:      host,
		port:      port,
		miners:    make(map[string]net.Conn),
		minerPass: make(map[string]string),
	}
}

// Start function
func (s *StratumServer) Start() {
	// the predefined username and password just for testing
	s.minerPass["Shuccck"] = "stratumServer"

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer ln.Close()
	log.Printf("Stratum server started on %s:%d\n", s.host, s.port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// connection handling of client side
func (s *StratumServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var request map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			log.Printf("Error decoding request: %v", err)
			continue
		}

		method, ok := request["method"].(string)
		if !ok {
			log.Println("Invalid method in request")
			continue
		}

		switch method {
		case "mining.subscribe":
			s.handleSubscribe(conn, request)
		case "mining.authorize":
			s.handleAuthorize(conn, request)
		case "mining.submit":
			s.handleSubmit(conn, request)
		default:
			log.Printf("Unsupported method: %s", method)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}

func (s *StratumServer) handleSubscribe(conn net.Conn, request map[string]interface{}) {
	response := map[string]interface{}{
		"id":     request["id"],
		"result": []interface{}{"mining.notify", "0x12345678", 30},
	}
	s.sendResponse(conn, response)
}

func (s *StratumServer) handleAuthorize(conn net.Conn, request map[string]interface{}) {
	params, ok := request["params"].([]interface{})
	if !ok || len(params) < 2 {
		s.sendError(conn, request["id"], "Invalid params")
		return
	}
	minerID, ok := params[0].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid miner ID")
		return
	}
	password, ok := params[1].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid password")
		return
	}

	expectedPassword, exists := s.minerPass[minerID]
	if !exists || password != expectedPassword {
		s.sendError(conn, request["id"], "Unauthorized")
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.miners[minerID] = conn
	s.sendResponse(conn, map[string]interface{}{"id": request["id"], "result": true})
}

// submission
func (s *StratumServer) handleSubmit(conn net.Conn, request map[string]interface{}) {
	params, ok := request["params"].([]interface{})
	if !ok || len(params) < 4 {
		s.sendError(conn, request["id"], "Invalid params")
		return
	}
	minerID, ok := params[0].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid miner ID")
		return
	}
	jobID, ok := params[1].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid job ID")
		return
	}
	nonce, ok := params[2].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid nonce")
		return
	}
	result, ok := params[3].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid result")
		return
	}

	log.Printf("Received solution from miner %s for job %s: nonce=%s, result=%s\n", minerID, jobID, nonce, result)
	s.sendResponse(conn, map[string]interface{}{"id": request["id"], "result": true})
}

func (s *StratumServer) sendResponse(conn net.Conn, response map[string]interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func (s *StratumServer) sendError(conn net.Conn, id interface{}, message string) {
	response := map[string]interface{}{
		"id":    id,
		"error": message,
	}
	s.sendResponse(conn, response)
}

func main() {
	server := NewStratumServer("localhost", 3333)
	server.Start()
}
