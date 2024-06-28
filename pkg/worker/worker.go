package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type Miner struct {
	ID     string
	Server string
	Port   int
	Conn   net.Conn
	Jobs   chan Job
}

func NewMiner(id, server string, port int) *Miner {
	return &Miner{
		ID:     id,
		Server: server,
		Port:   port,
		Jobs:   make(chan Job, 10),
	}
}

//connects to the Stratum server.
func (m *Miner) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.Server, m.Port))
	if err != nil {
		return err
	}
	m.Conn = conn
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
		"params": []interface{}{m.ID, "stratumServer"},
	}
	return m.sendRequest(request)
}

func (m *Miner) Start() {
	go m.receiveJobs()
	for job := range m.Jobs {
		m.mine(job)
	}
}

func (m *Miner) receiveJobs() {
	buffer := make([]byte, 4096)
	for {
		n, err := m.Conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from server: %v", err)
			close(m.Jobs)
			return
		}

		var job Job
		err = json.Unmarshal(buffer[:n], &job)
		if err != nil {
			log.Printf("Error unmarshalling job: %v", err)
			continue
		}

		m.Jobs <- job
	}
}

func (m *Miner) mine(job Job) {
	var wg sync.WaitGroup
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			nonce := rand.Uint32()
			for {
				select {
				case <-time.After(10 * time.Second):
					hash := sha256.Sum256([]byte(job.Data + string(nonce)))
					if isHashValid(hash) {
						result := hex.EncodeToString(hash[:])
						m.Submit(job.ID, string(nonce), result)
						return
					}
					nonce++
				}
			}
		}(i)
	}
	wg.Wait()
}

func isHashValid(hash [32]byte) bool {
	target := "0000"
	return hex.EncodeToString(hash[:])[:len(target)] == target
}

//submits mined solution.
func (m *Miner) Submit(jobID, nonce, result string) error {
	request := map[string]interface{}{
		"id":     3,
		"method": "mining.submit",
		"params": []interface{}{m.ID, jobID, nonce, result},
	}
	return m.sendRequest(request)
}

// send request sends request to the server.
func (m *Miner) sendRequest(request map[string]interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_, err = m.Conn.Write(data)
	return err
}

func main() {
	miner := NewMiner("Shuccck", "localhost", 3333)
	err := miner.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer miner.Conn.Close()

	err = miner.Subscribe()
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	err = miner.Authorize()
	if err != nil {
		log.Fatalf("Error authorizing: %v", err)
	}

	miner.Start()
}
