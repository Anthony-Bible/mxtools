// Package orchestration provides functionality for orchestrating and managing concurrent checks.
package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mxclone/pkg/types"
)

// CheckHandler is a function that handles a specific type of check.
type CheckHandler func(context.Context, *types.CheckRequest) types.Result

// Engine is the core orchestrator for the MXToolbox clone.
type Engine struct {
	workerPool  *WorkerPool
	handlers    map[string]CheckHandler
	jobsMu      sync.RWMutex
	jobs        map[string]*types.Job
	handlersMu  sync.RWMutex
}

// NewEngine creates a new engine with the specified number of workers.
func NewEngine(workerCount int) *Engine {
	e := &Engine{
		workerPool: NewWorkerPool(workerCount),
		handlers:   make(map[string]CheckHandler),
		jobs:       make(map[string]*types.Job),
	}
	e.workerPool.Start()
	go e.processResults()
	return e
}

// RegisterHandler registers a handler for a specific check type.
func (e *Engine) RegisterHandler(checkType string, handler CheckHandler) {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()
	e.handlers[checkType] = handler
}

// SubmitCheck submits a check request to the engine.
func (e *Engine) SubmitCheck(req types.CheckRequest) (string, error) {
	// Create a unique job ID
	jobID := fmt.Sprintf("%s-%d", req.Target, time.Now().UnixNano())
	
	// Create a job
	job := &types.Job{
		ID:      jobID,
		Request: req,
		Done:    false,
	}
	
	// Store the job
	e.jobsMu.Lock()
	e.jobs[jobID] = job
	e.jobsMu.Unlock()
	
	// Submit the job to the worker pool
	e.workerPool.Submit(job)
	
	return jobID, nil
}

// GetJobStatus returns the status of a job.
func (e *Engine) GetJobStatus(jobID string) (*types.Job, error) {
	e.jobsMu.RLock()
	defer e.jobsMu.RUnlock()
	
	job, ok := e.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	return job, nil
}

// processResults processes the results from the worker pool.
func (e *Engine) processResults() {
	for job := range e.workerPool.Results() {
		e.jobsMu.Lock()
		e.jobs[job.ID] = job
		e.jobsMu.Unlock()
	}
}

// Stop stops the engine.
func (e *Engine) Stop() {
	e.workerPool.Stop()
}

// ExecuteCheck executes a check request synchronously.
func (e *Engine) ExecuteCheck(ctx context.Context, req types.CheckRequest) types.Result {
	result := types.Result{
		Success: true,
		Message: "Check completed successfully",
	}
	
	// If no specific check types are requested, use all registered handlers
	checkTypes := req.CheckTypes
	if len(checkTypes) == 0 {
		e.handlersMu.RLock()
		for checkType := range e.handlers {
			checkTypes = append(checkTypes, checkType)
		}
		e.handlersMu.RUnlock()
	}
	
	// Execute each check type
	for _, checkType := range checkTypes {
		e.handlersMu.RLock()
		handler, ok := e.handlers[checkType]
		e.handlersMu.RUnlock()
		
		if !ok {
			result.Success = false
			result.Error = fmt.Sprintf("No handler registered for check type: %s", checkType)
			return result
		}
		
		// Execute the handler
		checkResult := handler(ctx, &req)
		
		// If any check fails, mark the overall result as failed
		if !checkResult.Success {
			result.Success = false
			result.Error = fmt.Sprintf("Check type %s failed: %s", checkType, checkResult.Error)
		}
		
		// Store the check result in the data map
		if result.Data == nil {
			result.Data = make(map[string]interface{})
		}
		result.Data.(map[string]interface{})[checkType] = checkResult.Data
	}
	
	return result
}