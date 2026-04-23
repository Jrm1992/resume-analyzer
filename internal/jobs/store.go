package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewStore() *Store {
	return &Store{jobs: make(map[string]*Job)}
}

func (s *Store) Create(resume, jd, language, requestID string) *Job {
	now := time.Now()
	j := &Job{
		ID:        uuid.NewString(),
		RequestID: requestID,
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		Resume:    resume,
		JD:        jd,
		Language:  language,
	}
	s.mu.Lock()
	s.jobs[j.ID] = j
	s.mu.Unlock()
	return j
}

// Get returns a shallow copy of the job.
//
// Invariant: Job.Result is a pointer to llm.AnalysisResult. Callers MUST treat
// Result as immutable — it is written exactly once (when the job transitions
// to StatusDone) and read by many. No deep copy is performed because the
// struct is not mutated after assignment; copying it would be pure overhead.
func (s *Store) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}

func (s *Store) Update(id string, mut func(*Job)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return
	}
	mut(j)
	j.UpdatedAt = time.Now()
}

func (s *Store) DeleteOlderThan(d time.Duration) int {
	cutoff := time.Now().Add(-d)
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, j := range s.jobs {
		if j.CreatedAt.Before(cutoff) {
			delete(s.jobs, id)
			n++
		}
	}
	return n
}
