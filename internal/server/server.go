package server

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// BlockHeader represents the block header structure.
type BlockHeader struct {
	Version        uint32 // Block version
	PrevBlockHash  string // Hash of the previous block
	MerkleRootHash string // Merkle root hash of transactions
	Timestamp      uint32 // Block timestamp
	Nonce          uint32 // Nonce used for mining
}

// Job represents a mining job.
type Job struct {
	ID   string `json:"id"`   // Job ID
	Data string `json:"data"` // Job data (block header in hexadecimal format)
}

var (
	jobCounter uint64 // Atomic counter for generating job IDs
)

// NewJob creates a new mining job with the provided block header.
func NewJob(header BlockHeader) *Job {
	// Serialize block header to hexadecimal format
	data, err := json.Marshal(header)
	if err != nil {
		log.Printf("Error encoding block header: %v", err)
		return nil
	}
	hexData := hex.EncodeToString(data)

	return &Job{
		ID:   generateJobID(), // Generate unique job ID
		Data: hexData,         // Set data field as hexadecimal representation of block header
	}
}

// GenerateBlockHeader generates a new block header for mining.
func GenerateBlockHeader(prevBlockHash string) BlockHeader {
	return BlockHeader{
		Version:        1,                         // Example: Set block version
		PrevBlockHash:  prevBlockHash,             // Set previous block hash
		MerkleRootHash: "dummy-merkle-root-hash",  // Example: Set Merkle root hash of transactions
		Timestamp:      uint32(time.Now().Unix()), // Set current timestamp
		Nonce:          0,                         // Initialize nonce to 0
	}
}

// NonceIsValid checks if the given nonce produces a valid block hash based on the target difficulty.
func NonceIsValid(header BlockHeader, targetDifficulty uint32) bool {
	// Serialize block header to hexadecimal format
	data, err := json.Marshal(header)
	if err != nil {
		log.Printf("Error encoding block header: %v", err)
		return false
	}
	hexData := hex.EncodeToString(data)

	// Concatenate nonce to block header
	headerWithNonce := hexData + intToHex(header.Nonce)

	// Calculate double SHA-256 hash of concatenated data
	hash := sha256.Sum256([]byte(headerWithNonce))
	hashString := hex.EncodeToString(hash[:])

	// Check if hash meets the target difficulty
	return hashString[:targetDifficulty/4] == "0000" // Example: Check if hash starts with four leading zeros
}

// intToHex converts an integer to a hexadecimal string.
func intToHex(n uint32) string {
	return hex.EncodeToString([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
}

// generateJobID generates a unique job ID.
func generateJobID() string {
	// Generate a unique job ID using a combination of timestamp and an atomic counter
	id := atomic.AddUint64(&jobCounter, 1)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("job-%d-%d", timestamp, id)
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
	solutionChannel   chan Solution // Channel to provide updated solutions
}

// NewStratumServer creates a new Stratum server.
func NewStratumServer(host string, port int, network string, connectionTimeout time.Duration, maxWorkers int) *StratumServer {
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
		solutionChannel:   make(chan Solution, 100), // Buffered channel with capacity 100
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

	// Send solution to the solution channel
	s.SendSolution(s.solutions[minerID])
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

// SendJob sends a job to a miner based on the miner's ID.
func (s *StratumServer) SendJob(job Job, minerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, ok := s.miners[minerID]
	if !ok {
		log.Printf("Miner with ID %s not found", minerID)
		return
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		log.Printf("Error encoding job data: %v", err)
		return
	}

	message := map[string]interface{}{
		"id":     "job",
		"method": "mining.notify",
		"params": []interface{}{jobData},
	}

	if err := json.NewEncoder(conn).Encode(message); err != nil {
		log.Printf("Error sending job to miner %s: %v", minerID, err)
		return
	}
}

// SetDifficulty sets the difficulty for a specific miner.
func (s *StratumServer) SetDifficulty(bits int, minerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if the miner exists
	conn, ok := s.miners[minerID]
	if !ok {
		log.Printf("Miner with ID %s not found", minerID)
		return
	}

	// Construct a response to set the difficulty
	response := map[string]interface{}{
		"id":     "set_difficulty",
		"method": "mining.set_difficulty",
		"params": []interface{}{bits},
	}

	// Encode and send the response to the miner
	if err := json.NewEncoder(conn).Encode(response); err != nil {
		log.Printf("Error setting difficulty for miner %s: %v", minerID, err)
		return
	}
}

// GetMinersLength returns the number of connected hardware miners.
func (s *StratumServer) GetMinersLength() uint32 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return uint32(len(s.miners))
}

// SendSolution sends a solution to the solution channel.
func (s *StratumServer) SendSolution(solution Solution) {
	s.solutionChannel <- solution
}
