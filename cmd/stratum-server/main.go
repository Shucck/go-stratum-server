package main

import (
	"difficulty"
	"server"
	"time"
)

func main() {
	server := server.NewStratumServer("localhost", 3333, "default_network", 5*time.Minute, 5, 1000) // Timeout set to 5 minutes, 5 workers, initial difficulty 1000
	server.Start()                                                                                  // Start the Stratum server

	// Using difficulty.go
	da := difficulty.NewDifficultyAdjuster()

	// to use
	da.AdjustDifficulty()
}
