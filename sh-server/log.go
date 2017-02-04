package main

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/unixpickle/statushub"
)

// Log maintains a history of statushub.LogRecords.
type Log struct {
	config     *Config
	logLock    sync.RWMutex
	curID      int
	perService map[string][]statushub.LogRecord
	allRecords []statushub.LogRecord

	serviceChans map[string]chan struct{}
	globalChan   chan struct{}
}

// NewLog creates a log which depends on a configuration
// to get the maximum log size.
func NewLog(cfg *Config) *Log {
	return &Log{
		config:       cfg,
		perService:   map[string][]statushub.LogRecord{},
		serviceChans: map[string]chan struct{}{},
	}
}

// Add adds a log record to the log.
func (l *Log) Add(service, msg string) (int, error) {
	ls := l.config.LogSize()

	// Technically, there is a possible scenario when
	// LogSizeUpdated() is called while we are right here,
	// resulting in a log which is larger than the current
	// log size.
	// However, this is not that big an issue and avoids
	// locking the configuration (which might be I/O bound)
	// while holding the log lock.

	l.logLock.Lock()
	record := statushub.LogRecord{
		Service: service,
		Message: msg,
		Time:    time.Now().Unix(),
		ID:      l.curID,
	}
	l.curID++
	l.allRecords = trimLog(append(l.allRecords, record), ls)
	l.perService[service] = trimLog(append(l.perService[service], record), ls)
	l.wakeListeners(service)
	l.logLock.Unlock()
	return record.ID, nil
}

// DeleteService deletes a service.
// It fails if the service does not exist.
func (l *Log) DeleteService(name string) error {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	if _, ok := l.perService[name]; !ok {
		return errors.New("no such service: " + name)
	}
	delete(l.perService, name)
	newLen := 0
	for _, x := range l.allRecords {
		if x.Service != name {
			l.allRecords[newLen] = x
			newLen++
		}
	}
	l.allRecords = l.allRecords[:newLen]
	l.wakeListeners(name)
	return nil
}

// Overview returns the most recent log record per
// service, sorted from most to least recent.
func (l *Log) Overview() []statushub.LogRecord {
	l.logLock.RLock()
	var entries []statushub.LogRecord
	for _, v := range l.perService {
		entries = append(entries, v[len(v)-1])
	}
	l.logLock.RUnlock()
	sort.Sort(logIDSorter(entries))
	return entries
}

// FullLog returns all of the log records, sorted from
// most to least recent.
func (l *Log) FullLog() []statushub.LogRecord {
	l.logLock.RLock()
	res := append([]statushub.LogRecord{}, l.allRecords...)
	l.logLock.RUnlock()
	return reverseLog(res)
}

// ServiceLog returns the log records for a particular
// service, sorted from most to least recent.
// It fails if there are no log records for the service.
func (l *Log) ServiceLog(name string) ([]statushub.LogRecord, error) {
	l.logLock.RLock()
	defer l.logLock.RUnlock()
	entries, ok := l.perService[name]
	if !ok {
		return nil, errors.New("unknown service: " + name)
	}
	return reverseLog(entries), nil
}

// LogSizeUpdated directs the log to delete log records as
// needed to accommodate the new log size.
func (l *Log) LogSizeUpdated() {
	ls := l.config.LogSize()
	l.logLock.Lock()
	l.allRecords = trimLog(l.allRecords, ls)
	for k, v := range l.perService {
		l.perService[k] = trimLog(v, ls)
	}
	l.logLock.Unlock()
}

// Wait creates a channel which is closed when any log
// entry is added or deleted.
// If the cancel chan is closed, the returned channel
// is closed early.
func (l *Log) Wait() <-chan struct{} {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	if l.globalChan == nil {
		l.globalChan = make(chan struct{})
	}
	return l.globalChan
}

// WaitService creates a channel which is closed when a
// log entry is added to the service, or when the service
// is deleted.
func (l *Log) WaitService(name string) <-chan struct{} {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	ch, ok := l.serviceChans[name]
	if !ok {
		ch = make(chan struct{})
		l.serviceChans[name] = ch
	}
	return ch
}

// wakeListeners wakes all the listeners for the service,
// as well as all global listeners.
//
// You should only call this while holding the log lock.
func (l *Log) wakeListeners(service string) {
	if ch, ok := l.serviceChans[service]; ok {
		close(ch)
		delete(l.serviceChans, service)
	}
	if l.globalChan != nil {
		close(l.globalChan)
		l.globalChan = nil
	}
}

func trimLog(log []statushub.LogRecord, maxSize int) []statushub.LogRecord {
	if maxSize == 0 {
		return log
	}
	if len(log) <= maxSize {
		return log
	}
	overflow := len(log) - maxSize
	copy(log[:], log[overflow:])
	return log[:maxSize]
}

func reverseLog(log []statushub.LogRecord) []statushub.LogRecord {
	res := make([]statushub.LogRecord, len(log))
	for i, x := range log {
		res[len(res)-(i+1)] = x
	}
	return res
}

type logIDSorter []statushub.LogRecord

func (l logIDSorter) Len() int {
	return len(l)
}

func (l logIDSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l logIDSorter) Less(i, j int) bool {
	return l[i].ID > l[j].ID
}
