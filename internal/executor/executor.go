package executor

import (
	"context"

	"github.com/Igordro1d/job_scheduler/internal/job"
)

type HandlerFunc func(ctx context.Context, j *job.Job) error

type Registry struct {
	handlers map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

func (r *Registry) Register(jobType string, h HandlerFunc) {
	r.handlers[jobType] = h
}

func (r *Registry) Get(jobType string) (HandlerFunc, bool) {
	h, ok := r.handlers[jobType]
	return h, ok
}
