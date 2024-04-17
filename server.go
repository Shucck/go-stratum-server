package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Job represents a mining job.
type Job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// Solution represents a mining solution.
type Solution struct {
	MinerID string `json:"minerId"`
	JobID   string `json:"jobId"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

// StratumServer represents a Stratum mining server.
type StratumServer struct {
	host      string
	port      int
	networks  map[string]map[string]net.Conn
	jobs      map[string]Job
	solutions map[string]Solution
	mutex     sync.Mutex
}

// NewStratumServer creates a new Stratum server.
func NewStratumServer(host string, port int) *StratumServer {
	return &StratumServer{
		host:      host,
		port:      port,
		networks:  make(map[string]map[string]net.Conn),
		jobs:      make(map[string]Job),
		solutions: make(map[string]Solution),
	}
}

// Start starts the Stratum server.
func (s *StratumServer) Start() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Printf("Stratum server started on %s:%d\n", s.host, s.port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection handles a client connection.
func (s *StratumServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var request map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			fmt.Println("Error decoding request:", err)
			continue
		}

		method, ok := request["method"].(string)
		if !ok {
			fmt.Println("Invalid method in request")
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
			fmt.Println("Unsupported method:", method)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

// handleSubscribe handles a mining subscription request.
func (s *StratumServer) handleSubscribe(conn net.Conn, request map[string]interface{}) {
	response := map[string]interface{}{
		"id":     request["id"],
		"result": []interface{}{"mining.notify", "0x12345678", 30},
	}
	s.sendResponse(conn, response)
}

// handleAuthorize handles a mining authorization request.
func (s *StratumServer) handleAuthorize(conn net.Conn, request map[string]interface{}) {
	params, ok := request["params"].([]interface{})
	if !ok || len(params) == 0 {
		s.sendError(conn, request["id"], "Invalid params")
		return
	}
	minerID, ok := params[0].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid miner ID")
		return
	}

	// Assume network information is provided in params[1]
	network, ok := params[1].(string)
	if !ok {
		s.sendError(conn, request["id"], "Invalid network")
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.networks[network]; !ok {
		s.networks[network] = make(map[string]net.Conn)
	}
	s.networks[network][minerID] = conn

	s.sendResponse(conn, map[string]interface{}{"id": request["id"], "result": true})
}

// handleSubmit handles a mining submission request.
func (s *StratumServer) handleSubmit(conn net.Conn, request map[string]interface{}) {
	params, ok := request["params"].([]interface{})
	if !ok || len(params) != 4 {
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.solutions[minerID]; !ok {
		s.solutions[minerID] = Solution{}
	}
	s.solutions[minerID] = Solution{MinerID: minerID, JobID: jobID, Nonce: nonce, Result: result}

	s.sendResponse(conn, map[string]interface{}{"id": request["id"], "result": true})
}

// sendResponse sends a response to the client.
func (s *StratumServer) sendResponse(conn net.Conn, response map[string]interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error encoding response:", err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending response:", err)
	}
}

// sendError sends an error response to the client.
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
