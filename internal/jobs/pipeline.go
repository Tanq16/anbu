package jobs

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
	"uuid"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"

	finishedCap = 50
	jobTimeout  = 30 * time.Minute
)

var ErrNotFound = errors.New("job not found")

type Job struct {
	ID       string    `json:"id"`
	Kind     string    `json:"kind"`
	Label    string    `json:"label"`
	Status   string    `json:"status"`
	Result   any       `json:"result,omitempty"`
	Error    string    `json:"error,omitempty"`
	Queued   time.Time `json:"queued"`
	Started  time.Time `json:"started,omitzero"`
	Finished time.Time `json:"finished,omitzero"`
	run      func(ctx context.Context) (any, error)
}

type Snapshot struct {
	Running  *Job  `json:"running"`
	Queued   []Job `json:"queued"`
	Finished []Job `json:"finished"`
}

type Pipeline struct {
	mu       sync.Mutex
	queue    []*Job
	current  *Job
	finished []*Job
	wake     chan struct{}
}

func New() *Pipeline {
	p := &Pipeline{wake: make(chan struct{}, 1)}
	go p.worker()
	return p
}

func (p *Pipeline) Enqueue(kind, label string, run func(ctx context.Context) (any, error)) Job {
	job := &Job{
		ID:     uuid.NewV7().String(),
		Kind:   kind,
		Label:  label,
		Status: StatusQueued,
		Queued: time.Now().UTC(),
		run:    run,
	}
	p.mu.Lock()
	p.queue = append(p.queue, job)
	snapshot := *job
	p.mu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
	return snapshot
}

func (p *Pipeline) worker() {
	for range p.wake {
		for job := p.next(); job != nil; job = p.next() {
			result, err := runJob(job)
			p.finish(job, result, err)
		}
	}
}

func runJob(job *Job) (result any, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job panicked: %v", r)
		}
	}()
	return job.run(ctx)
}

func (p *Pipeline) next() *Job {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.queue) == 0 {
		return nil
	}
	job := p.queue[0]
	p.queue = slices.Delete(p.queue, 0, 1)
	job.Status = StatusRunning
	job.Started = time.Now().UTC()
	p.current = job
	return job
}

func (p *Pipeline) finish(job *Job, result any, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	job.Finished = time.Now().UTC()
	job.Result = result
	job.Status = StatusSucceeded
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()
	}
	p.current = nil
	p.finished = slices.Insert(p.finished, 0, job)
	if len(p.finished) > finishedCap {
		p.finished = p.finished[:finishedCap]
	}
}

func (p *Pipeline) Snapshot() Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	s := Snapshot{Queued: copyJobs(p.queue), Finished: copyJobs(p.finished)}
	if p.current != nil {
		running := *p.current
		s.Running = &running
	}
	return s
}

func (p *Pipeline) Get(id string) (Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current != nil && p.current.ID == id {
		return *p.current, nil
	}
	for _, list := range [][]*Job{p.queue, p.finished} {
		if i := slices.IndexFunc(list, func(j *Job) bool { return j.ID == id }); i >= 0 {
			return *list[i], nil
		}
	}
	return Job{}, ErrNotFound
}

func copyJobs(jobs []*Job) []Job {
	out := make([]Job, len(jobs))
	for i, j := range jobs {
		out[i] = *j
	}
	return out
}
