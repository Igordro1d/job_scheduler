package api

import (
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/job"
	"github.com/Igordro1d/job_scheduler/internal/store"
	"github.com/jackc/pgx/v5"
)

//go:embed web/index.html
var webFS embed.FS

type Server struct {
	store store.Store
}

func NewServer(s store.Store) *Server {
	return &Server{store: s}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /jobs", s.handleEnqueue)
	mux.HandleFunc("GET /jobs", s.handleListJobs)
	mux.HandleFunc("GET /jobs/{id}", s.handleGetJob)
	mux.HandleFunc("GET /dead-letter", s.handleListDeadLetter)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	page, err := webFS.ReadFile("web/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(page)
}

type enqueueRequest struct {
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	Priority       int             `json:"priority"`
	DependsOn      *string         `json:"depends_on"`
	IdempotencyKey *string         `json:"idempotency_key"`
}

type jobResponse struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Priority   int       `json:"priority"`
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	DependsOn  *string   `json:"depends_on"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toJobResponse(j *job.Job) jobResponse {
	return jobResponse{
		ID:         j.ID,
		Type:       j.Type,
		Status:     string(j.Status),
		Priority:   j.Priority,
		RetryCount: j.RetryCount,
		MaxRetries: j.MaxRetries,
		DependsOn:  j.DependsOn,
		CreatedAt:  j.CreatedAt,
		UpdatedAt:  j.UpdatedAt,
	}
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	var req enqueueRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	created, err := s.store.Enqueue(r.Context(), job.EnqueueParams{
		Type:           req.Type,
		Payload:        req.Payload,
		Priority:       req.Priority,
		DependsOn:      req.DependsOn,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toJobResponse(created))
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	j, err := s.store.GetByID(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(j))
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.store.ListRecent(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	responses := make([]jobResponse, 0, len(jobs))
	for _, j := range jobs {
		responses = append(responses, toJobResponse(j))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (s *Server) handleListDeadLetter(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.store.ListDeadLetter(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	responses := make([]jobResponse, 0, len(jobs))
	for _, j := range jobs {
		responses = append(responses, toJobResponse(j))
	}

	writeJSON(w, http.StatusOK, responses)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
