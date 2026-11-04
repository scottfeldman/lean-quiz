package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type dayState struct {
	Answers   [][]string `json:"answers"`
	Submitted bool       `json:"submitted"`
}

type fileState struct {
	Days map[string]dayState `json:"days"`
}

type store struct {
	mu   sync.Mutex
	path string
	data fileState
}

func loadStore(path string) (*store, error) {
	s := &store{path: path, data: fileState{Days: map[string]dayState{}}}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, err
	}
	if s.data.Days == nil {
		s.data.Days = map[string]dayState{}
	}
	return s, nil
}

func (s *store) get(day int, item day) dayState {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.data.Days[strconv.Itoa(day)]
	if !ok || !fits(item, st) {
		return dayState{Answers: freshAnswers(item)}
	}
	out := dayState{Submitted: st.Submitted, Answers: make([][]string, len(st.Answers))}
	for i, row := range st.Answers {
		out.Answers[i] = append([]string(nil), row...)
	}
	return out
}

func fits(item day, st dayState) bool {
	if len(st.Answers) != len(item.questions) {
		return false
	}
	for i, q := range item.questions {
		if len(st.Answers[i]) != q.slots() {
			return false
		}
	}
	return true
}

func freshAnswers(item day) [][]string {
	out := make([][]string, len(item.questions))
	for i, q := range item.questions {
		out[i] = make([]string, q.slots())
	}
	return out
}

func (s *store) put(day int, st dayState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Days == nil {
		s.data.Days = map[string]dayState{}
	}
	s.data.Days[strconv.Itoa(day)] = st
	return s.save()
}

func (s *store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
