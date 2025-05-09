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

func (s *TracerouteJobStore) StartCleanup(expiry time.Duration, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			s.cleanupExpired(expiry)
		}
	}()
}

func (s *TracerouteJobStore) cleanupExpired(expiry time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, job := range s.jobs {
		if (job.Status == JobComplete || job.Status == JobError) && job.CompletedAt != nil {
			if now.Sub(*job.CompletedAt) > expiry {
				delete(s.jobs, id)
			}
		}
	}
}

func GetTracerouteJobStore() *TracerouteJobStore {
	return tracerouteJobStore
}

func init() {
	// Start cleanup with 10 min expiry, runs every 1 min
	tracerouteJobStore.StartCleanup(10*time.Minute, 1*time.Minute)
}
