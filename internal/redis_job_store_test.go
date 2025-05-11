package internal

import (
	"context"
	"encoding/json"
	"errors"
	"mxclone/pkg/redisiface"
	"testing"
	"time"
)

// --- Mock Redis Client for Unit Testing ---
// Implements redisiface.RedisClient
type mockRedisClient struct {
	store     map[string][]byte
	failSet   bool
	failGet   bool
	failWatch bool
}

// Ensure mockRedisClient implements redisiface.RedisClient
var _ redisiface.RedisClient = (*mockRedisClient)(nil)

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{store: make(map[string][]byte)}
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.failSet {
		return errors.New("mock set error")
	}
	b, _ := value.([]byte)
	m.store[key] = b
	return nil
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.failGet {
		return "", errors.New("mock get error")
	}
	v, ok := m.store[key]
	if !ok {
		return "", errors.New("not found")
	}
	return string(v), nil
}

func (m *mockRedisClient) Watch(ctx context.Context, fn func(redisiface.RedisClient) error, keys ...string) error {
	if m.failWatch {
		return errors.New("mock watch error")
	}
	return fn(m)
}

func (m *mockRedisClient) TxPipelined(ctx context.Context, fn func(redisiface.RedisClient) error) error {
	return fn(m)
}

func (m *mockRedisClient) Close() error { return nil }

// --- Unit Test Setup ---
func setupMockRedisJobStore() *RedisJobStore {
	mock := newMockRedisClient()
	return NewRedisJobStoreWithClient(mock, "testjob:")
}

// --- Unit Tests ---
func TestRedisJobStore_AddGet(t *testing.T) {
	store := setupMockRedisJobStore()
	job := &TracerouteJob{
		JobID:     "job1",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	b, _ := json.Marshal(job)
	err := store.client.Set(context.Background(), store.jobKey(job.JobID), b, 0)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	val, err := store.client.Get(context.Background(), store.jobKey(job.JobID))
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	var got TracerouteJob
	json.Unmarshal([]byte(val), &got)
	if got.JobID != job.JobID {
		t.Errorf("Got job %+v, want %+v", got, job)
	}
}

func TestRedisJobStore_Update(t *testing.T) {
	store := setupMockRedisJobStore()
	job := &TracerouteJob{
		JobID:     "job2",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	err := store.Add(job)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	err = store.Update("job2", func(j *TracerouteJob) {
		j.Status = "complete"
		completed := time.Now()
		j.CompletedAt = &completed
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	got, found, err := store.Get("job2")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if !found {
		t.Fatalf("Job not found after update")
	}
	if got.Status != "complete" {
		t.Errorf("Expected status 'complete', got '%s'", got.Status)
	}
	if got.CompletedAt == nil {
		t.Errorf("Expected CompletedAt to be set, got nil")
	}
}

func TestRedisJobStore_GetNotFound(t *testing.T) {
	store := setupMockRedisJobStore()
	_, found, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if found {
		t.Errorf("Expected not found for nonexistent job")
	}
}

func TestRedisJobStore_AddDuplicate(t *testing.T) {
	store := setupMockRedisJobStore()
	job := &TracerouteJob{
		JobID:     "jobdup",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	err := store.Add(job)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	// Add again with different status
	job.Status = "running"
	err = store.Add(job)
	if err != nil {
		t.Fatalf("Add duplicate failed: %v", err)
	}
	got, found, err := store.Get("jobdup")
	if err != nil {
		t.Fatalf("Get after duplicate add failed: %v", err)
	}
	if !found {
		t.Fatalf("Job not found after duplicate add")
	}
	if got.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", got.Status)
	}
}

func TestRedisJobStore_UpdateNotFound(t *testing.T) {
	store := setupMockRedisJobStore()
	err := store.Update("doesnotexist", func(j *TracerouteJob) {
		j.Status = "error"
	})
	if err == nil {
		t.Errorf("Expected error when updating nonexistent job, got nil")
	}
}

func TestRedisJobStore_Close(t *testing.T) {
	store := setupMockRedisJobStore()
	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestRedisJobStore_ConcurrentAddUpdate(t *testing.T) {
	store := setupMockRedisJobStore()
	jobID := "concurrentjob"
	job := &TracerouteJob{
		JobID:     jobID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	err := store.Add(job)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	n := 10
	done := make(chan bool, n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			err := store.Update(jobID, func(j *TracerouteJob) {
				j.Status = TracerouteJobStatus("update" + string(rune('A'+idx)))
			})
			if err != nil {
				t.Errorf("Concurrent update failed: %v", err)
			}
			done <- true
		}(i)
	}
	for i := 0; i < n; i++ {
		<-done
	}
	got, found, err := store.Get(jobID)
	if err != nil || !found {
		t.Fatalf("Get after concurrent updates failed: %v", err)
	}
	if got.Status == "pending" {
		t.Errorf("Expected status to be updated, got 'pending'")
	}
}

func TestRedisJobStore_SetError(t *testing.T) {
	mock := newMockRedisClient()
	mock.failSet = true
	store := NewRedisJobStoreWithClient(mock, "testjob:")
	job := &TracerouteJob{JobID: "errset", Status: "pending", CreatedAt: time.Now()}
	err := store.Add(job)
	if err == nil {
		t.Errorf("Expected error from Set, got nil")
	}
}

func TestRedisJobStore_GetError(t *testing.T) {
	mock := newMockRedisClient()
	mock.failGet = true
	store := NewRedisJobStoreWithClient(mock, "testjob:")
	_, found, err := store.Get("errget")
	if err == nil {
		t.Errorf("Expected error from Get, got nil")
	}
	if found {
		t.Errorf("Expected found=false on Get error")
	}
}

func TestRedisJobStore_WatchError(t *testing.T) {
	mock := newMockRedisClient()
	mock.failWatch = true
	store := NewRedisJobStoreWithClient(mock, "testjob:")
	err := store.Update("any", func(j *TracerouteJob) { j.Status = "fail" })
	if err == nil {
		t.Errorf("Expected error from Watch, got nil")
	}
}

func TestRedisJobStore_JSONUnmarshalError(t *testing.T) {
	mock := newMockRedisClient()
	store := NewRedisJobStoreWithClient(mock, "testjob:")
	// Insert invalid JSON
	mock.store[store.jobKey("badjson")] = []byte("notjson")
	_, found, err := store.Get("badjson")
	if err == nil {
		t.Errorf("Expected error from JSON unmarshal, got nil")
	}
	if found {
		t.Errorf("Expected found=false on JSON error")
	}
}

func TestRedisJobStore_StartCleanup_NoOp(t *testing.T) {
	store := setupMockRedisJobStore()
	// Should not panic or error
	store.StartCleanup(10*time.Minute, 1*time.Minute)
}
