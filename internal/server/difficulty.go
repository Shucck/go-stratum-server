package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// DifficultyAdjuster represents a difficulty adjuster for a Stratum mining server.
type DifficultyAdjuster struct {
	server         *StratumServer
	windowSize     int
	targetInterval time.Duration
	currentWindow  []float64
	mutex          sync.Mutex
}

// NewDifficultyAdjuster creates a new difficulty adjuster.
func NewDifficultyAdjuster(server *StratumServer, windowSize int, targetInterval time.Duration) *DifficultyAdjuster {
	return &DifficultyAdjuster{
		server:         server,
		windowSize:     windowSize,
		targetInterval: targetInterval,
		currentWindow:  make([]float64, 0, windowSize),
		mutex:          sync.Mutex{},
	}
}

// AdjustDifficulty adjusts the mining difficulty based on the share times.
func (da *DifficultyAdjuster) AdjustDifficulty(shareTime float64) {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	// Add share time to current window
	da.currentWindow = append(da.currentWindow, shareTime)

	// If window size exceeded, calculate new difficulty
	if len(da.currentWindow) >= da.windowSize {
		// Calculate difficulty adjustment...
		newDiff := calculateNewDifficulty(da.currentWindow, da.targetInterval)

		// Update the server's difficulty
		da.server.UpdateDifficulty(newDiff)

		// Send updated difficulty to server
		select {
		case da.server.DifficultyChannel <- newDiff:
		default:
			log.Println("Difficulty update skipped: channel full")
		}

		// Clear the current window
		da.currentWindow = make([]float64, 0, da.windowSize)
	}
}

// calculateNewDifficulty calculates the new difficulty based on share times and target interval.
func calculateNewDifficulty(window []float64, targetInterval time.Duration) int {
	// Calculate new difficulty...
	return 0 // Placeholder value, replace with actual calculation
}

// StratumServer represents a Stratum mining server.
type StratumServer struct {
	// Other fields...
	difficulty        int      // Current share difficulty
	DifficultyChannel chan int // Channel to broadcast updated difficulty
}

// NewStratumServer creates a new Stratum server.
func NewStratumServer(host string, port int, network string, connectionTimeout time.Duration, maxWorkers int, difficulty int) *StratumServer {
	return &StratumServer{
		// Initialize other fields...
		difficulty:        difficulty,
		DifficultyChannel: make(chan int),
	}
}

// UpdateDifficulty updates the server's current difficulty.
func (s *StratumServer) UpdateDifficulty(newDifficulty int) {
	s.difficulty = newDifficulty
}

// broadcastDifficulty sends the updated difficulty to all connected miners.
func (s *StratumServer) broadcastDifficulty() {
	for newDiff := range s.DifficultyChannel {
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
