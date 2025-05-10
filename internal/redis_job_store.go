// filepath: /home/anthony/GolandProjects/mxclone/internal/redis_job_store.go
package internal

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"mxclone/pkg/logging" // Assuming a logging package

	"github.com/redis/go-redis/v9"
)

// RedisJobStore implements the JobStore interface using Redis.
type RedisJobStore struct {
	client *redis.Client
	prefix string // Prefix for Redis keys to avoid collisions
}

// NewRedisJobStore creates a new RedisJobStore.
// addr is the Redis server address (e.g., "localhost:6379").
// password is the Redis password (empty if none).
// db is the Redis database number.
// prefix is a string to prefix all keys with (e.g., "traceroutejob:").
func NewRedisJobStore(addr, password string, db int, prefix string) (*RedisJobStore, error) {
	logging.Info("RedisJobStore: Creating Redis client with addr=%s, db=%d, prefix=%s", addr, db, prefix) // Password omitted for security
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	maxRetries := 5 // Number of times to retry connection
	retryDelay := 3 * time.Second

	var err error
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = rdb.Ping(ctx).Result()
		cancel() // Release resources associated with context

		if err == nil {
			// Connection successful
			return &RedisJobStore{
				client: rdb,
				prefix: prefix,
			}, nil
		}

		logging.Warn("RedisJobStore: Failed to connect to Redis (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	// All retries failed
	return nil, errors.New("failed to connect to Redis after multiple retries: " + err.Error())
}

func (s *RedisJobStore) jobKey(jobID string) string {
	return s.prefix + jobID
}

// Add adds a new job to Redis.
func (s *RedisJobStore) Add(job *TracerouteJob) error {
	ctx := context.Background()
	jobJSON, err := json.Marshal(job)
	if err != nil {
		logging.Error("RedisJobStore: Failed to marshal job for Add: %v", err)
		return err // Propagate error
	}

	// Store the job with no specific expiry here; expiry is handled by CompletedAt + TTL if needed,
	// or by a separate cleanup mechanism if jobs can be pending indefinitely.
	// For simplicity, we'll rely on the job's CompletedAt for cleanup logic if StartCleanup is called.
	err = s.client.Set(ctx, s.jobKey(job.JobID), jobJSON, 0).Err()
	if err != nil {
		logging.Error("RedisJobStore: Failed to add job to Redis: %v", err)
		return err
	}
	return nil // No error
}

// Get retrieves a job from Redis by its ID.
// It returns the job, true if found, and an error if any other issue occurred.
func (s *RedisJobStore) Get(jobID string) (*TracerouteJob, bool, error) {
	ctx := context.Background()
	val, err := s.client.Get(ctx, s.jobKey(jobID)).Result()
	if err == redis.Nil {
		return nil, false, nil // Job not found, no error
	} else if err != nil {
		logging.Error("RedisJobStore: Failed to get job from Redis: %v", err)
		return nil, false, err // Other error
	}

	var job TracerouteJob
	err = json.Unmarshal([]byte(val), &job)
	if err != nil {
		logging.Error("RedisJobStore: Failed to unmarshal job from Redis: %v", err)
		return nil, false, err // Unmarshal error
	}
	return &job, true, nil // Job found
}

// Update modifies an existing job in Redis.
// It uses a WATCH/MULTI/EXEC transaction for atomic updates.
func (s *RedisJobStore) Update(jobID string, updateFn func(*TracerouteJob)) error {
	ctx := context.Background()
	key := s.jobKey(jobID)

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		val, err := tx.Get(ctx, key).Result()
		if err == redis.Nil {
			return errors.New("job not found for update")
		} else if err != nil {
			return err
		}

		var job TracerouteJob
		if err := json.Unmarshal([]byte(val), &job); err != nil {
			return err
		}

		updateFn(&job) // Apply the update

		jobJSON, err := json.Marshal(job)
		if err != nil {
			return err
		}

		// Use MULTI/EXEC to set the new value.
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, jobJSON, 0) // No expiry, or set based on job.CompletedAt
			return nil
		})
		return err
	}, key)

	if err != nil {
		logging.Error("RedisJobStore: Failed to update job in Redis: %v", err)
		return err
	}
	return nil // Update successful
}

// StartCleanup for RedisJobStore.
// Redis can handle TTLs automatically, so this might be a no-op or
// could implement a more complex scanning cleanup for jobs that don't have TTLs
// or where TTLs are managed based on job status and CompletedAt.
// For this example, we'll make it a no-op, assuming TTLs are set on Add/Update if desired,
// or a separate Redis-native cleanup (like key eviction policies) is in place.
// A more robust implementation might scan for completed jobs and set TTLs.
func (s *RedisJobStore) StartCleanup(expiry time.Duration, interval time.Duration) {
	// This is a simplified version. A production Redis store might:
	// 1. Set TTLs on jobs when they are marked complete/error.
	// 2. Have a separate process that scans for completed jobs and sets TTLs or deletes them.
	// For now, we log that this is a manual or TTL-based process for Redis.
	logging.Info("RedisJobStore: Cleanup for Redis is typically handled by TTLs set on keys or Redis eviction policies. Manual periodic cleanup not implemented in this version.")

	// Example of how one might set a TTL when a job is completed (would be in Update):
	// if job.Status == JobComplete || job.Status == JobError {
	//   s.client.Expire(ctx, key, expiry)
	// }
}

// Close closes the Redis client connection.
func (s *RedisJobStore) Close() error {
	return s.client.Close()
}
