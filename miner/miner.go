package main

import (
	"net"
	"time"
)

// Miner represents a connected miner.
type Miner struct {
	ID            string
	Connection    net.Conn
	LastActivity  time.Time
	Subscription  string
	Difficulty    int
	Authorized    bool
}
