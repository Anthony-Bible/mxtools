// set of redis.Client used by RedisJobStore (for mocking in tests)
package internal

import (
	"context"
	"encoding/json"
	"errors"
	"mxclone/pkg/logging"
	"mxclone/pkg/redisiface"
	"time"
)

type redisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Watch(ctx context.Context, fn func(redisClient) error, keys ...string) error
	TxPipelined(ctx context.Context, fn func(redisClient) error) error
	Close() error
}

// RedisJobStore implements the JobStore interface using redisiface.RedisClient.
type RedisJobStore struct {
	client redisiface.RedisClient
	prefix string // Prefix for Redis keys to avoid collisions
}

// NewRedisJobStoreWithClient allows injecting a mock RedisClient (for unit tests).
func NewRedisJobStoreWithClient(client redisiface.RedisClient, prefix string) *RedisJobStore {
	return &RedisJobStore{
		client: client,
		prefix: prefix,
	}
}

func (s *RedisJobStore) jobKey(jobID string) string {
	return s.prefix + jobID
}

// Add adds a new job to Redis.
// It returns an error if the job could not be added.
func (s *RedisJobStore) Add(job *TracerouteJob) error {
	ctx := context.Background()
	jobJSON, err := json.Marshal(job)
	if err != nil {
		logging.Error("RedisJobStore: Failed to marshal job for Add: %v", err)
		return err // Propagate error
	}
	return s.client.Set(ctx, s.jobKey(job.JobID), jobJSON, 0)
}

// Get retrieves a job from Redis by its ID.
// It returns the job, true if found, and an error if any other issue occurred.
func (s *RedisJobStore) Get(jobID string) (*TracerouteJob, bool, error) {
	ctx := context.Background()
	val, err := s.client.Get(ctx, s.jobKey(jobID))
	if err != nil {
		if err.Error() == "not found" {
			return nil, false, nil // Job not found, no error
		}
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
// It uses a transaction for atomic updates.
func (s *RedisJobStore) Update(jobID string, updateFn func(*TracerouteJob)) error {
	ctx := context.Background()
	key := s.jobKey(jobID)

	err := s.client.Watch(ctx, func(tx redisiface.RedisClient) error {
		val, err := tx.Get(ctx, key)
		if err != nil {
			return errors.New("job not found for update")
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

		// Use TxPipelined to set the new value.
		// Assuming tx.TxPipelined also expects a function taking redisiface.RedisClient or a compatible pipeliner type.
		// If redisiface.RedisClient's TxPipelined provides a different type for `pipe`, adjust accordingly.
		// For now, we'll assume it's consistent with Watch.
		err = tx.TxPipelined(ctx, func(pipe redisiface.RedisClient) error {
			return pipe.Set(ctx, key, jobJSON, 0) // No expiry, or set based on job.CompletedAt
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
