package testutil

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

// StartServer starts the server
func StartServer(t *testing.T) *net.TCPListener {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Error starting server: %v", err)
	}
	return l.(*net.TCPListener)
}

// ConnectServer connects to the server
func ConnectServer(t *testing.T, l *net.TCPListener) *net.TCPConn {
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	return conn.(*net.TCPConn)
}

// SendRequest sends a request to the server
func SendRequest(t *testing.T, conn *net.TCPConn, request map[string]interface{}) {
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Error encoding message: %v", err)
	}
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("Error sending message: %v", err)
	}
}

// WaitShort waits for a short duration
func WaitShort() {
	time.Sleep(100 * time.Millisecond)
}
