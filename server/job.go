package main

// Job represents a mining job.
type Job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// NewJob creates a new job with a unique ID.
func NewJob(data string) Job {
	id := generateJobID()
	return Job{
		ID:   id,
		Data: data,
	}
}

// generateJobID generates a unique job ID.
func generateJobID() string {
	return fmt.Sprintf("%08x", atomic.AddUint64(&jobCounter, 1))
}
