package server

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	TOTAL_CONNECTIONS string = "perf-total-connections"
	QUIC_CONN         string = "perf-quic-from-%s"
)

var Statistics = &statistics{}

func init() {
	Statistics.Reset()
}

type statistics struct {
	semCounters   *sync.RWMutex
	semAddressMap *sync.RWMutex

	counters        map[string]int
	sourceToDestMap map[string]string
}

func (s *statistics) init() {
	if s.semCounters != nil && s.semAddressMap != nil {
		return
	}

	s.semCounters = &sync.RWMutex{}
	s.semAddressMap = &sync.RWMutex{}
}

func (s *statistics) Reset() {
	s.semCounters = nil
	s.semAddressMap = nil
	s.init()

	log.Println("Server statistics reset.")
	s.counters = make(map[string]int)
	s.sourceToDestMap = make(map[string]string)
}

// ---- Counters ---- //
func (s *statistics) Get(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}

	s.init()
	s.semCounters.RLock()
	defer s.semCounters.RUnlock()

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

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	s.counters[key] = value
	return value
}

func (s *statistics) Increment(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		log.Printf("counter: %s = -1\n", key)
		return -1
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	value, ok := s.counters[key]
	if !ok {
		log.Printf("counter: %s = *%d\n", key, 1)
		s.counters[key] = 1
		return 1
	}

	log.Printf("counter: %s = %d\n", key, value+1)
	s.counters[key] = value + 1
	return value + 1
}

func (s *statistics) Decrement(key string) int {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) == 0 {
		return -1
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	value, ok := s.counters[key]
	if !ok || value-1 <= 0 {
		s.counters[key] = 0
		return 0
	}

	log.Printf("counter: %s = %d\n", key, value-1)
	s.counters[key] = value - 1
	return value - 1
}

// ---- address mapping ---- //
func (s *statistics) GetMappedAddress(source string) string {
	s.init()
	s.semAddressMap.RLock()
	defer s.semAddressMap.RUnlock()

	if val, ok := s.sourceToDestMap[source]; ok {
		return val
	}
	return ""
}

func (s *statistics) SetMappedAddress(source string, dest string) {
	s.init()
	s.semAddressMap.Lock()
	defer s.semAddressMap.Unlock()

	s.sourceToDestMap[source] = dest
}

func (s *statistics) DeleteMappedAddress(source string) {
	s.init()
	s.semAddressMap.Lock()
	defer s.semAddressMap.Unlock()

	delete(s.sourceToDestMap, source)
}
