// Package orchestration provides functionality for orchestrating and managing concurrent checks.
package orchestration

import (
	"context"
	"sync"

	"mxclone/pkg/types"
)

// WorkerPool manages a pool of workers for processing jobs.
type WorkerPool struct {
	workerCount int
	jobQueue    chan *types.Job
	results     chan *types.Job
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(workerCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan *types.Job, workerCount*2), // Buffer size is twice the worker count
		results:     make(chan *types.Job, workerCount*2),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker pool.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop stops the worker pool.
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.results)
}

// Submit submits a job to the worker pool.
func (wp *WorkerPool) Submit(job *types.Job) {
	select {
	case wp.jobQueue <- job:
		// Job submitted successfully
	case <-wp.ctx.Done():
		// Worker pool is shutting down
	}
}

// Results returns a channel that receives completed jobs.
func (wp *WorkerPool) Results() <-chan *types.Job {
	return wp.results
}

// worker is the worker goroutine that processes jobs.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				// Job queue is closed
				return
			}
			// Process the job
			wp.processJob(job)
			// Send the result
			select {
			case wp.results <- job:
				// Result sent successfully
			case <-wp.ctx.Done():
				// Worker pool is shutting down
				return
			}
		case <-wp.ctx.Done():
			// Worker pool is shutting down
			return
		}
	}
}

// processJob processes a job.
func (wp *WorkerPool) processJob(job *types.Job) {
	// This is a placeholder. In a real implementation, this would dispatch
	// to the appropriate handler based on the job type.
	// For now, we'll just set the job as done.
	job.Done = true
}