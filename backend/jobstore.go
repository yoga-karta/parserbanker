package main

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type JobStatus string

const (
	StatusStaging    JobStatus = "staging" // job dibuat, menunggu file di-load
	StatusLoaded     JobStatus = "loaded"  // kedua file sudah di-load, siap diproses
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusError      JobStatus = "error"
	StatusStopped    JobStatus = "stopped"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type Job struct {
	ID        string      `json:"id"`
	Status    JobStatus   `json:"status"`
	Error     string      `json:"error,omitempty"`
	Summary   Summary     `json:"summary,omitempty"`
	DBPath    string      `json:"-"`
	EJPath    string      `json:"-"`
	RCPath    string      `json:"-"`
	EJLoaded  bool        `json:"ej_loaded"`
	RCLoaded  bool        `json:"rc_loaded"`
	EJRecords int         `json:"ej_records"`
	RCRecords int         `json:"rc_records"`
	Options   LoadOptions `json:"options"`
	Log       []LogEntry  `json:"log"`
	cancel    context.CancelFunc
}

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

func (s *JobStore) Create(id string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := &Job{ID: id, Status: StatusStaging}
	s.jobs[id] = j
	s.appendLogLocked(j, "Job dibuat, menunggu file EJ dan RC")
	return j
}

// Get mengembalikan pointer internal ke job - HANYA aman dipakai untuk cek
// keberadaan (ok bool) atau diteruskan langsung ke method JobStore lain yang
// sudah locking sendiri. Jangan baca field dari hasil Get() di luar lock -
// pakai Snapshot() untuk itu.
func (s *JobStore) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

// Snapshot mengembalikan SALINAN (bukan pointer) dari state job saat ini,
// termasuk salinan independen dari slice Log. Aman dibaca, dikirim ke goroutine
// lain, atau langsung di-JSON-marshal tanpa risiko data race dengan
// SetDone/SetError/AppendLog dkk yang jalan concurrent di goroutine lain.
func (s *JobStore) Snapshot(id string) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return Job{}, false
	}
	cp := *j
	cp.Log = append([]LogEntry(nil), j.Log...)
	cp.cancel = nil
	return cp, true
}

func (s *JobStore) appendLogLocked(j *Job, msg string) {
	j.Log = append(j.Log, LogEntry{Timestamp: time.Now(), Message: msg})
}

func (s *JobStore) AppendLog(id, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		s.appendLogLocked(j, msg)
	}
}

func (s *JobStore) SetEJLoaded(id, path string, records int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.EJPath = path
		j.EJRecords = records
		j.EJLoaded = true
		if j.RCLoaded {
			j.Status = StatusLoaded
		}
		s.appendLogLocked(j, "File EJ berhasil di-load: "+strconv.Itoa(records)+" record")
	}
}

func (s *JobStore) SetRCLoaded(id, path string, records int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.RCPath = path
		j.RCRecords = records
		j.RCLoaded = true
		if j.EJLoaded {
			j.Status = StatusLoaded
		}
		s.appendLogLocked(j, "File RC berhasil di-load: "+strconv.Itoa(records)+" record")
	}
}

func (s *JobStore) SetOptions(id string, opts LoadOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Options = opts
	}
}

func (s *JobStore) SetProcessing(id string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = StatusProcessing
		j.cancel = cancel
		s.appendLogLocked(j, "Proses rekonsiliasi dimulai")
	}
}

func (s *JobStore) SetDone(id string, summary Summary, dbPath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = StatusDone
		j.Summary = summary
		j.DBPath = dbPath
		j.cancel = nil
		s.appendLogLocked(j, "Proses selesai")
	}
}

func (s *JobStore) SetError(id string, err string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = StatusError
		j.Error = err
		j.cancel = nil
		s.appendLogLocked(j, "Error: "+err)
	}
}

// Stop membatalkan job yang sedang diproses. Mengembalikan false kalau job
// tidak sedang berjalan (tidak ada apa-apa buat dibatalkan).
func (s *JobStore) Stop(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok || j.cancel == nil {
		return false
	}
	j.cancel()
	j.Status = StatusStopped
	s.appendLogLocked(j, "Proses dihentikan oleh user (STOP)")
	return true
}
