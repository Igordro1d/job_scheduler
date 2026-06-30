package job

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusDead       Status = "dead"
)

type Job struct {
	ID             string
	Type           string
	Payload        json.RawMessage
	Priority       int
	Status         Status
	DependsOn      *string
	MaxRetries     int
	RetryCount     int
	IdempotencyKey *string
	LockedBy       *string
	LockedAt       *time.Time
	RunAfter       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EnqueueParams struct {
	Type           string
	Payload        json.RawMessage
	Priority       int
	DependsOn      *string
	IdempotencyKey *string
}
