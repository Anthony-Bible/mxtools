package internal

import (
	"os"
	"testing"
	"time"
)

func setupTestRedisJobStore(t *testing.T) *RedisJobStore {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	store, err := NewRedisJobStore(addr, "", 0, "testjob:")
	if err != nil {
		t.Fatalf("Failed to create RedisJobStore: %v", err)
	}
	return store
}

func TestRedisJobStore_AddGet(t *testing.T) {
	store := setupTestRedisJobStore(t)
	job := &TracerouteJob{
		JobID:     "job1",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	err := store.Add(job)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	got, found, err := store.Get("job1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Fatalf("Job not found after Add")
	}
	if got.JobID != job.JobID {
		t.Errorf("Got job %+v, want %+v", got, job)
	}
}

func TestRedisJobStore_Update(t *testing.T) {
	store := setupTestRedisJobStore(t)
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
	store := setupTestRedisJobStore(t)
	_, found, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if found {
		t.Errorf("Expected not found for nonexistent job")
	}
}

func TestRedisJobStore_AddDuplicate(t *testing.T) {
	store := setupTestRedisJobStore(t)
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
	store := setupTestRedisJobStore(t)
	err := store.Update("doesnotexist", func(j *TracerouteJob) {
		j.Status = "error"
	})
	if err == nil {
		t.Errorf("Expected error when updating nonexistent job, got nil")
	}
}

func TestRedisJobStore_Close(t *testing.T) {
	store := setupTestRedisJobStore(t)
	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
