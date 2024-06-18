package main

// Solution represents a solution submitted by a miner.
type Solution struct {
	MinerID   string
	JobID     string
	Nonce     string
	Result    string
	Timestamp int64
}

// Worker represents a worker that processes mining jobs.
type Worker struct {
	ID              string
	SolutionChannel chan Solution
	StopChannel     chan bool
}

// NewWorker creates a new worker.
func NewWorker(id string, solutionChannel chan Solution) Worker {
	return Worker{
		ID:              id,
		SolutionChannel: solutionChannel,
		StopChannel:     make(chan bool),
	}
}

// Start starts the worker to process solutions.
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case solution := <-w.SolutionChannel:
				w.processSolution(solution)
			case <-w.StopChannel:
				return
			}
		}
	}()
}

// Stop stops the worker from processing solutions.
func (w *Worker) Stop() {
	w.StopChannel <- true
}

// processSolution processes a solution submitted by a miner.
func (w *Worker) processSolution(solution Solution) {
	// Process the solution, for example, validate it and store it.
	log.Printf("Worker %s processing solution from miner %s for job %s\n", w.ID, solution.MinerID, solution.JobID)
}
