package state

import (
	"sync"
)

// Storage provides thread-safe in-memory key-value storage
type Storage struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		store: make(map[string][]byte),
	}
}

// Set stores a value for the given key
func (s *Storage) Set(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
	return nil
}

// Get retrieves a value for the given key
func (s *Storage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.store[key]
	if !exists {
		return nil, nil
	}
	return value, nil
}

// Has checks if a key exists
func (s *Storage) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.store[key]
	return exists
}

// GetAllKeys returns all keys in the storage (for debugging/status)
func (s *Storage) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys
}

