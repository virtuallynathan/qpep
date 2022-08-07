package api

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	TOTAL_CONNECTIONS string = "perf-total-connections"
	QUIC_CONN         string = "perf-quic-from"
	QUIC_HOSTS        string = "perf-host"
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
	hosts           []string
}

func (s *statistics) init() {
	if s.semCounters != nil && s.semAddressMap != nil {
		return
	}

	s.semCounters = &sync.RWMutex{}
	s.semAddressMap = &sync.RWMutex{}
	s.hosts = make([]string, 0, 32)
}

func (s *statistics) Reset() {
	s.semCounters = nil
	s.semAddressMap = nil
	s.init()

	log.Println("Server statistics reset.")
	s.counters = make(map[string]int)
	s.sourceToDestMap = make(map[string]string)
}

func (s *statistics) AsKey(prefix string, values ...string) string {
	return strings.ToLower(prefix + "-" + strings.Join(values, "-"))
}

// ---- Counters ---- //
func (s *statistics) Get(key string) int {
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

	if len(key) == 0 {
		return -1
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	s.counters[key] = value
	return value
}

func (s *statistics) Increment(prefix string, keyparts ...string) int {

	key := s.AsKey(prefix, keyparts...)
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

func (s *statistics) Decrement(prefix string, keyparts ...string) int {
	key := s.AsKey(prefix, keyparts...)
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

	if _, ok := s.sourceToDestMap[source]; !ok {
		s.hosts = append(s.hosts, dest)
	}
	s.sourceToDestMap[source] = dest
}

func (s *statistics) DeleteMappedAddress(source string) {
	s.init()
	s.semAddressMap.Lock()
	defer s.semAddressMap.Unlock()

	if _, ok := s.sourceToDestMap[source]; ok {
		mapped := s.sourceToDestMap[source]
		for i := 0; i < len(s.hosts); i++ {
			if !strings.EqualFold(s.hosts[i], mapped) {
				continue
			}
			s.hosts = append(s.hosts[:i], s.hosts[i+1:]...)
			break
		}
	}
	delete(s.sourceToDestMap, source)
}

// ---- hosts ---- //
func (s *statistics) GetHosts() []string {
	s.init()
	s.semAddressMap.RLock()
	defer s.semAddressMap.RUnlock()

	return append([]string{}, s.hosts...)
}
