package tasks

import (
	"cmp"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
	"uuid"

	"github.com/tanq16/anbu/internal/datadir"
)

var (
	ErrNotFound = errors.New("task not found")
	ErrInvalid  = errors.New("invalid task")
)

type Priority string

const (
	PriorityDefault  Priority = "default"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

var priorityRank = map[Priority]int{PriorityCritical: 0, PriorityHigh: 1, PriorityDefault: 2}

type Task struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Priority  Priority  `json:"priority"`
	Due       string    `json:"due,omitempty"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at,omitzero"`
}

type Input struct {
	Text     string   `json:"text"`
	Priority Priority `json:"priority"`
	Due      string   `json:"due"`
	Done     bool     `json:"done"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	tasks []Task
}

func Open(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.tasks); err != nil {
		return nil, err
	}
	return s, nil
}

func (in *Input) normalize() error {
	in.Text = strings.TrimSpace(in.Text)
	in.Due = strings.TrimSpace(in.Due)
	if in.Text == "" {
		return fmt.Errorf("%w: text is required", ErrInvalid)
	}
	if strings.ContainsAny(in.Text, "\r\n") {
		return fmt.Errorf("%w: text must be a single line", ErrInvalid)
	}
	if in.Priority == "" {
		in.Priority = PriorityDefault
	}
	if _, ok := priorityRank[in.Priority]; !ok {
		return fmt.Errorf("%w: priority must be default, high, or critical", ErrInvalid)
	}
	if in.Due != "" {
		if _, err := time.Parse(time.DateOnly, in.Due); err != nil {
			return fmt.Errorf("%w: due must be a YYYY-MM-DD date", ErrInvalid)
		}
	}
	return nil
}

func (s *Store) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := slices.Clone(s.tasks)
	slices.SortFunc(out, compare)
	return out
}

func compare(a, b Task) int {
	if a.Done != b.Done {
		if a.Done {
			return 1
		}
		return -1
	}
	if a.Done {
		return cmp.Or(b.DoneAt.Compare(a.DoneAt), a.CreatedAt.Compare(b.CreatedAt))
	}
	if c := cmp.Compare(priorityRank[a.Priority], priorityRank[b.Priority]); c != 0 {
		return c
	}
	if a.Due != b.Due {
		switch {
		case a.Due == "":
			return 1
		case b.Due == "":
			return -1
		}
		return cmp.Compare(a.Due, b.Due)
	}
	return a.CreatedAt.Compare(b.CreatedAt)
}

func (s *Store) Create(in Input) (Task, error) {
	if err := in.normalize(); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{
		ID:        uuid.NewV7().String(),
		Text:      in.Text,
		Priority:  in.Priority,
		Due:       in.Due,
		CreatedAt: time.Now().UTC(),
	}
	s.tasks = append(s.tasks, t)
	if err := s.persist(); err != nil {
		s.tasks = s.tasks[:len(s.tasks)-1]
		return Task{}, err
	}
	return t, nil
}

func (s *Store) Update(id string, in Input) (Task, error) {
	if err := in.normalize(); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.index(id)
	if i < 0 {
		return Task{}, ErrNotFound
	}
	prev := s.tasks[i]
	t := prev
	t.Text, t.Priority, t.Due = in.Text, in.Priority, in.Due
	switch {
	case in.Done && !prev.Done:
		t.DoneAt = time.Now().UTC()
	case !in.Done:
		t.DoneAt = time.Time{}
	}
	t.Done = in.Done
	s.tasks[i] = t
	if err := s.persist(); err != nil {
		s.tasks[i] = prev
		return Task{}, err
	}
	return t, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.index(id)
	if i < 0 {
		return ErrNotFound
	}
	prev := slices.Clone(s.tasks)
	s.tasks = slices.Delete(s.tasks, i, i+1)
	if err := s.persist(); err != nil {
		s.tasks = prev
		return err
	}
	return nil
}

func (s *Store) index(id string) int {
	return slices.IndexFunc(s.tasks, func(t Task) bool { return t.ID == id })
}

func (s *Store) persist() error {
	data, err := json.Marshal(s.tasks)
	if err != nil {
		return err
	}
	return datadir.WriteFile(s.path, data)
}
