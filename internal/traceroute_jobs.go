package internal

import (
	"sync"
	"time"
	"github.com/google/uuid"
	"mxclone/domain/networktools"
)

type TracerouteJobStatus string

const (
	JobPending   TracerouteJobStatus = "pending"
	JobRunning   TracerouteJobStatus = "running"
	JobComplete  TracerouteJobStatus = "complete"
	JobError     TracerouteJobStatus = "error"
)

type TracerouteJob struct {
	JobID      string               `json:"jobId"`
	Status     TracerouteJobStatus  `json:"status"`
	Result     *networktools.TracerouteResult `json:"result,omitempty"`
	Error      string               `json:"error,omitempty"`
	CreatedAt  time.Time            `json:"createdAt"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
}

type TracerouteJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*TracerouteJob
}

var tracerouteJobStore = &TracerouteJobStore{
	jobs: make(map[string]*TracerouteJob),
}

func NewTracerouteJob(host string) *TracerouteJob {
	return &TracerouteJob{
		JobID:     uuid.NewString(),
		Status:    JobPending,
		CreatedAt: time.Now(),
	}
}

func (s *TracerouteJobStore) Add(job *TracerouteJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.JobID] = job
}

func (s *TracerouteJobStore) Get(jobID string) (*TracerouteJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	return job, ok
}

func (s *TracerouteJobStore) Update(jobID string, update func(*TracerouteJob)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[jobID]; ok {
		update(job)
	}
}

func GetTracerouteJobStore() *TracerouteJobStore {
	return tracerouteJobStore
}
