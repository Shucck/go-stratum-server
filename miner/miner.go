package server

import (
	"net"
	"time"
)

// Miner represents a connected miner.
type Miner struct {
	ID         string
	Connection net.Conn
	LastActive time.Time
}

// NewMiner creates a new Miner instance.
func NewMiner(id string, conn net.Conn) *Miner {
	return &Miner{
		ID:         id,
		Connection: conn,
		LastActive: time.Now(),
	}
}

// SetLastActive sets the last active time for the miner.
func (m *Miner) SetLastActive() {
	m.LastActive = time.Now()
}

// IsInactive checks if the miner is inactive based on the specified duration.
func (m *Miner) IsInactive(duration time.Duration) bool {
	return time.Since(m.LastActive) > duration
}
