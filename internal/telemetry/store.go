package telemetry

import (
	"sync"
	"time"
)

type Event struct {
	ID        string         `json:"id"`
	VideoID   string         `json:"video_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Type      string         `json:"type"`
	Position  float64        `json:"position,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	At        time.Time      `json:"at"`
}

type Session struct {
	ID        string    `json:"id"`
	VideoID   string    `json:"video_id"`
	UserID    string    `json:"user_id,omitempty"`
	StartedAt time.Time `json:"started_at"`
	LastSeen  time.Time `json:"last_seen"`
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]Session
	events   []Event
}

func NewStore() *Store { return &Store{sessions: make(map[string]Session)} }

func (s *Store) AddSession(session Session) Session {
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	return session
}

func (s *Store) GetSession(id string) (Session, bool) {
	s.mu.RLock()
	v, ok := s.sessions[id]
	s.mu.RUnlock()
	return v, ok
}

func (s *Store) AddEvent(event Event) {
	s.mu.Lock()
	s.events = append(s.events, event)
	if session, ok := s.sessions[event.SessionID]; ok {
		session.LastSeen = event.At
		s.sessions[event.SessionID] = session
	}
	s.mu.Unlock()
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}
