package api

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	TOTAL_CONNECTIONS  string = "counter-connections" // total connections open on the server at this time
	PERF_CONN          string = "perf-connections"    // number of current connections for a particular client
	PERF_UP_COUNT      string = "perf-up-count"       // current upload speed for a particular client
	PERF_DW_COUNT      string = "perf-dw-count"       // current download speed for a particular client
	PERF_UP_SPEED      string = "perf-up-speed"       // current upload speed for a particular client
	PERF_DW_SPEED      string = "perf-dw-speed"       // current download speed for a particular client
	PERF_UP_TOTAL      string = "perf-up-total"       // total number of bytes uploaded by a particular client
	PERF_DW_TOTAL      string = "perf-dw-total"       // total number of bytes downloaded by a particular client
	INFO_PLATFORM      string = "info-platform"       // platform used by the client, as communicated in api echo
	INFO_UPDATE        string = "info-update"         // last time server received an echo from the client
	INFO_OTHER_VERSION string = "info-remote-version" // version of the software on the other end of the connection
)

var Statistics = &statistics{}

func init() {
	Statistics.Reset()
}

type statistics struct {
	semCounters *sync.RWMutex
	semState    *sync.RWMutex

	counters map[string]float64
	state    map[string]string
	hosts    []string
}

func (s *statistics) init() {
	if s.semCounters != nil && s.semState != nil {
		return
	}

	log.Println("Statistics init.")
	s.semCounters = &sync.RWMutex{}
	s.semState = &sync.RWMutex{}
	s.hosts = make([]string, 0, 32)
}

func (s *statistics) Reset() {
	s.semCounters = nil
	s.semState = nil
	s.init()

	log.Println("Statistics reset.")
	s.counters = make(map[string]float64)
	s.state = make(map[string]string)
}

func (s *statistics) asKey(prefix string, values ...string) string {
	if len(values) == 0 {
		return strings.ToLower(prefix) + "[]"
	}
	return strings.ToLower(prefix + "[" + strings.Join(values, "-") + "]")
}

// ---- Counters ---- //
func (s *statistics) GetCounter(prefix string, keyparts ...string) float64 {
	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return -1
	}

	s.init()
	s.semCounters.RLock()
	defer s.semCounters.RUnlock()

	//log.Printf("GET counter: %s = %.2f\n", key, s.counters[key])
	if val, ok := s.counters[key]; ok {
		return val
	}
	return -1
}

func (s *statistics) SetCounter(value float64, prefix string, keyparts ...string) float64 {
	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return -1
	}
	if value < 0 {
		panic(fmt.Sprintf("Will not track negative values: (%s: %.2f)", key, value))
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	s.counters[key] = value
	//log.Printf("SET counter: %s = %.2f\n", key, s.counters[key])
	return value
}

func (s *statistics) GetCounterAndClear(prefix string, keyparts ...string) float64 {
	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return -1
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	//log.Printf("GET+CLEAR counter: %s = %.2f\n", key, s.counters[key])
	if val, ok := s.counters[key]; ok {
		s.counters[key] = 0.0
		return val
	}
	return -1
}

func (s *statistics) IncrementCounter(incr float64, prefix string, keyparts ...string) float64 {
	if incr < 0.0 {
		panic("Cannot increase value by a negative value!")
	}

	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return -1
	}
	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	value, ok := s.counters[key]
	if !ok {
		s.counters[key] = incr
		return incr
	}

	s.counters[key] = value + incr
	return s.counters[key]
}

func (s *statistics) DecrementCounter(decr float64, prefix string, keyparts ...string) float64 {
	if decr < 0.0 {
		panic("Please specify decrement as a positive value!")
	}

	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return -1.0
	}

	s.init()
	s.semCounters.Lock()
	defer s.semCounters.Unlock()

	value, ok := s.counters[key]
	if !ok || value-1.0 <= 0.0 {
		s.counters[key] = 0.0
		return 0.0
	}

	//log.Printf("counter: %s = %.2f\n", key, value-decr)
	s.counters[key] = value - decr
	return value - decr
}

// ---- State ---- //
func (s *statistics) GetState(prefix string, keyparts ...string) string {
	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return ""
	}

	s.init()
	s.semState.RLock()
	defer s.semState.RUnlock()

	if val, ok := s.state[key]; ok {
		return val
	}
	return ""
}

func (s *statistics) SetState(value, prefix string, keyparts ...string) string {
	key := s.asKey(prefix, keyparts...)
	if len(key) == 0 {
		return ""
	}

	s.init()
	s.semState.Lock()
	defer s.semState.Unlock()

	s.state[key] = value
	return value
}

// ---- address mapping ---- //
func (s *statistics) GetMappedAddress(source string) string {
	s.init()
	s.semState.RLock()
	defer s.semState.RUnlock()

	if val, ok := s.state[source]; ok {
		return val
	}
	return ""
}

func (s *statistics) SetMappedAddress(source string, dest string) {
	s.init()
	s.semState.Lock()
	defer s.semState.Unlock()

	if _, ok := s.state[source]; !ok {
		found := false
		for i := 0; i < len(s.hosts); i++ {
			if strings.EqualFold(s.hosts[i], dest) {
				found = true
				break
			}
		}
		if !found {
			s.hosts = append(s.hosts, dest)
		}
	}
	s.state[source] = dest
}

func (s *statistics) DeleteMappedAddress(source string) {
	s.init()
	s.semState.Lock()
	defer s.semState.Unlock()

	if _, ok := s.state[source]; ok {
		mapped := s.state[source]
		for i := 0; i < len(s.hosts); i++ {
			if !strings.EqualFold(s.hosts[i], mapped) {
				continue
			}
			s.hosts = append(s.hosts[:i], s.hosts[i+1:]...)
			break
		}
	}
	delete(s.state, source)
}

// ---- hosts ---- //
func (s *statistics) GetHosts() []string {
	s.init()
	s.semState.RLock()
	defer s.semState.RUnlock()

	// for test
	//log.Printf("hosts: %v\n", strings.Join(s.hosts, ","))
	//v := append([]string{}, "127.0.0.1")
	v := append([]string{}, s.hosts...)
	return v
}
