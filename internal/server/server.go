package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Job represents a mining job.
type Job struct {
	ID         string `json:"id"`
	Data       string `json:"data"`
	Difficulty int    `json:"difficulty"`
}

// Solution represents a mining solution.
type Solution struct {
	MinerID   string `json:"minerId"`
	JobID     string `json:"jobId"`
	Nonce     string `json:"nonce"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

// StratumServer represents a Stratum mining server.
type StratumServer struct {
	host              string
	port              int
	network           string // Network for all miners
	miners            map[string]net.Conn
	jobQueue          []Job
	solutions         map[string]Solution
	mutex             sync.Mutex
	connectionTimeout time.Duration // Connection timeout duration
	workerPool        chan struct{} // Worker pool for job processing
	maxWorkers        int           // Maximum number of worker goroutines
	difficulty        int           // Current share difficulty
	difficultyChannel chan int      // Channel to broadcast updated difficulty
}

// NewStratumServer creates a new Stratum server.
func NewStratumServer(host string, port int, network string, connectionTimeout time.Duration, maxWorkers int, difficulty int) *StratumServer {
	return &StratumServer{
		host:              host,
		port:              port,
		network:           network,
		miners:            make(map[string]net.Conn),
		jobQueue:          []Job{},
		solutions:         make(map[string]Solution),
		mutex:             sync.Mutex{},
		connectionTimeout: connectionTimeout,
		workerPool:        make(chan struct{}, maxWorkers),
		difficulty:        difficulty,
		difficultyChannel: make(chan int),
	}
}

// Start starts the Stratum server.
func (s *StratumServer) Start() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer ln.Close()
	log.Printf("Stratum server started on %s:%d\n", s.host, s.port)

	// Start a goroutine to check for connection timeouts
	go s.checkConnectionTimeouts()

	// Start a goroutine to broadcast difficulty updates
	go s.broadcastDifficulty()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// checkConnectionTimeouts checks for idle connections and closes them.
func (s *StratumServer) checkConnectionTimeouts() {
	for {
		time.Sleep(s.connectionTimeout)
		s.mutex.Lock()
		for minerID, conn := range s.miners {
			lastActivity := time.Now().Sub(conn.(*net.TCPConn).RemoteAddr().(*net.TCPAddr).Timestamp)
			if lastActivity > s.connectionTimeout {
				conn.Close()
				delete(s.miners, minerID)
				log.Printf("Closed idle connection for miner %s\n", minerID)
			}
		}
		s.mutex.Unlock()
	}
}

// handleConnection handles a client connection.
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.miners[minerID] = conn

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

	// Check if miner is authorized
	if _, ok := s.miners[minerID]; !ok {
		s.sendError(conn, request["id"], "Miner is not authorized")
		return
	}

	// Process the submission
	if _, ok := s.solutions[minerID]; !ok {
		s.solutions[minerID] = Solution{}
	}
	s.solutions[minerID] = Solution{MinerID: minerID, JobID: jobID, Nonce: nonce, Result: result, Timestamp: time.Now().Unix()}

	s.sendResponse(conn, map[string]interface{}{"id": request["id"], "result": true})
}

// sendResponse sends a response to the client.
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

// sendError sends an error response to the client.
func (s *StratumServer) sendError(conn net.Conn, id interface{}, message string) {
	response := map[string]interface{}{
		"id":    id,
		"error": message,
	}
	s.sendResponse(conn, response)
}

// broadcastDifficulty sends the updated difficulty to all connected miners.
func (s *StratumServer) broadcastDifficulty() {
	for newDiff := range s.difficultyChannel {
		// Construct difficulty update message
		response := map[string]interface{}{
			"method": "mining.set_difficulty",
			"params": []interface{}{newDiff},
		}

		// Broadcast difficulty update to all miners
		s.mutex.Lock()
		for _, conn := range s.miners {
			if err := json.NewEncoder(conn).Encode(response); err != nil {
				log.Printf("Error broadcasting difficulty update: %v", err)
			}
		}
		s.mutex.Unlock()
	}
}
