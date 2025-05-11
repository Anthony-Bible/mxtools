package internal

import (
	"errors"
	"log/slog" // Changed from mxclone/logging to log/slog
	"mxclone/adapters/secondary"
	"mxclone/internal/config"
	"sync"
	"time"

	"mxclone/domain/networktools"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TracerouteJobStatus string

const (
	JobPending  TracerouteJobStatus = "pending"
	JobRunning  TracerouteJobStatus = "running"
	JobComplete TracerouteJobStatus = "complete"
	JobError    TracerouteJobStatus = "error"
)

type TracerouteJob struct {
	JobID       string                         `json:"jobId"`
	Status      TracerouteJobStatus            `json:"status"`
	Result      *networktools.TracerouteResult `json:"result,omitempty"`
	Error       string                         `json:"error,omitempty"`
	CreatedAt   time.Time                      `json:"createdAt"`
	CompletedAt *time.Time                     `json:"completedAt,omitempty"`
}

// JobStore defines the interface for storing and managing traceroute jobs.
type JobStore interface {
	Add(job *TracerouteJob) error
	Get(jobID string) (*TracerouteJob, bool, error) // Added error return
	Update(jobID string, update func(*TracerouteJob)) error
	StartCleanup(expiry time.Duration, interval time.Duration)
}

// InMemoryJobStore implements the JobStore interface using an in-memory map.
type InMemoryJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*TracerouteJob
}

var globalJobStore JobStore // Use the interface type

func NewTracerouteJob(host string) *TracerouteJob {
	return &TracerouteJob{
		JobID:     uuid.NewString(),
		Status:    JobPending,
		CreatedAt: time.Now(),
	}
}

// Add adds a new job to the store.
func (s *InMemoryJobStore) Add(job *TracerouteJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.JobID] = job
	return nil
}

// Get retrieves a job from the store by its ID.
// It returns the job, true if found, and an error (always nil for InMemoryJobStore).
func (s *InMemoryJobStore) Get(jobID string) (*TracerouteJob, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	return job, ok, nil // InMemoryJobStore Get operation itself doesn't produce errors beyond not found
}

// Update modifies an existing job in the store.
func (s *InMemoryJobStore) Update(jobID string, update func(*TracerouteJob)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[jobID]; ok {
		update(job)
	} else {
		return errors.New("job not found for update")
	}
	return nil
}

// StartCleanup starts a background goroutine to periodically remove expired jobs.
// This method is specific to InMemoryJobStore and might be handled differently
// by other JobStore implementations (e.g., Redis TTL).
func (s *InMemoryJobStore) StartCleanup(expiry time.Duration, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			s.cleanupExpired(expiry)
		}
	}()
}

func (s *InMemoryJobStore) cleanupExpired(expiry time.Duration) {
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

// GetJobStore returns the global job store instance.
// It now returns the JobStore interface.
func GetJobStore() JobStore {
	return globalJobStore
}

// InitJobStore initializes the job store based on configuration
func InitJobStore(cfg *config.Config) {
	if cfg.JobStoreType == "redis" {
		slog.Info("Redis address configuration", "address", cfg.Redis.Address)
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		adapter := &secondary.RedisClientAdapter{Client: rdb}
		redisStore := NewRedisJobStoreWithClient(adapter, cfg.Redis.Prefix)
		globalJobStore = redisStore
		// StartCleanup for RedisJobStore is a no-op or handled by Redis TTLs
	} else {
		slog.Info("Using InMemoryJobStore")
		inMemoryStore := &InMemoryJobStore{
			jobs: make(map[string]*TracerouteJob),
		}
		inMemoryStore.StartCleanup(10*time.Minute, 1*time.Minute)
		globalJobStore = inMemoryStore
	}
}

// Initialize with default InMemoryJobStore for safety
func init() {
	// Default to in-memory store at package init time
	inMemoryStore := &InMemoryJobStore{
		jobs: make(map[string]*TracerouteJob),
	}
	inMemoryStore.StartCleanup(10*time.Minute, 1*time.Minute)
	globalJobStore = inMemoryStore
}
