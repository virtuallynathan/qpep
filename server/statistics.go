package server

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	TOTAL_CONNECTIONS string = "srv-total-connections"
	QUIC_CONN         string = "srv-quic-from-%s"
)

var Statistics = &statistics{}

func init() {
	Statistics.Reset()
}

type statistics struct {
	sem sync.RWMutex

	counters map[string]int
}

func (s *statistics) Reset() {
	s.sem.Lock()
	defer s.sem.Unlock()

	log.Println("Server statistics reset.")
	s.counters = make(map[string]int)
}

func (s *statistics) Get(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}
	s.sem.RLock()
	defer s.sem.RUnlock()

	if val, ok := s.counters[key]; ok {
		return val
	}
	return -1
}

func (s *statistics) Set(key string, value int) int {
	if value < 0 {
		panic(fmt.Sprintf("Will not track negative values: (%s: %d)", key, value))
	}

	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}
	s.sem.Lock()
	defer s.sem.Unlock()

	s.counters[key] = value
	return value
}

func (s *statistics) Increment(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}
	s.sem.Lock()
	defer s.sem.Unlock()

	value, ok := s.counters[key]
	if !ok {
		s.counters[key] = 1
		return 1
	}

	s.counters[key] = value + 1
	return value + 1
}

func (s *statistics) Decrement(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}
	s.sem.Lock()
	defer s.sem.Unlock()

	value, ok := s.counters[key]
	if !ok || value-1 <= 0 {
		s.counters[key] = 0
		return 0
	}

	s.counters[key] = value - 1
	return value - 1
}
