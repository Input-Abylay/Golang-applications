package store

import (
	"sort"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]Task
}

func New() *Store {
	return &Store{
		nextID: 1,
		tasks:  make(map[int]Task),
	}
}

func (s *Store) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t
}

func (s *Store) Get(id int) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Store) List(doneFilter *bool) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if doneFilter != nil && t.Done != *doneFilter {
			continue
		}
		out = append(out, t)
	}

	// stable order
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) UpdateDone(id int, done bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return false
	}
	t.Done = done
	s.tasks[id] = t
	return true
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}
