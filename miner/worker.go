package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// Job represents a mining job.
type Job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

var (
	jobCounter uint64 // Atomic counter for generating job IDs
)

// NewJob creates a new mining job with the provided block header.
package server

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "sync/atomic"
    "time"
)

// Define the BlockHeader type here
type BlockHeader struct {
    // Define the fields of the BlockHeader
}

// Job represents a mining job.
type Job struct {
    ID   string `json:"id"`
    Data string `json:"data"`
}

var (
    jobCounter uint64 // Atomic counter for generating job IDs
)

// generateJobID generates a unique job ID.
func generateJobID() string {
    id := atomic.AddUint64(&jobCounter, 1)
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("job-%d-%d", timestamp, id)
}

// NewJob creates a new mining job with the provided block header.
func NewJob(header BlockHeader) *Job {
    data, err := json.Marshal(header)
    if err != nil {
        log.Printf("Error encoding block header: %v", err)
        return nil
    }
    hexData := hex.EncodeToString(data)

    return &Job{
        ID:   generateJobID(),
        Data: hexData,
    }
}

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "sync/atomic"
    "time"
)

// Define the BlockHeader type here if it's not defined in another package
// type BlockHeader struct {
// 	// Define the fields of the BlockHeader
// }

// Job represents a mining job.
type Job struct {
    ID   string `json:"id"`
    Data string `json:"data"`
}

var (
    jobCounter uint64 // Atomic counter for generating job IDs
)

// NewJob creates a new mining job with the provided block header.
func NewJob(header BlockHeader) *Job {
    data, err := json.Marshal(header)
    if err!= nil {
        log.Printf("Error encoding block header: %v", err)
        return nil
    }
    hexData := hex.EncodeToString(data)

    return &Job{
        ID:   generateJobID(),
        Data: hexData,
    }
}

// generateJobID generates a unique job ID.
func generateJobID() string {
    id := atomic.AddUint64(&jobCounter, 1)
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("job-%d-%d", timestamp, id)
}
	data, err := json.Marshal(header)
	if err != nil {
		log.Printf("Error encoding block header: %v", err)
		return nil
	}
	hexData := hex.EncodeToString(data)

	return &Job{
		ID:   generateJobID(),
		Data: hexData,
	}
}

// generateJobID generates a unique job ID.
func generateJobID() string {
	id := atomic.AddUint64(&jobCounter, 1)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("job-%d-%d", timestamp, id)
}
