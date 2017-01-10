package main

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// A LogRecord is one logged message.
type LogRecord struct {
	Service string `json:"serviceName"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
	ID      int    `json:"id"`
}

// Log maintains a history of LogRecords.
type Log struct {
	config     *Config
	logLock    sync.RWMutex
	curID      int
	perService map[string][]LogRecord
	allRecords []LogRecord
}

func NewLog(cfg *Config) *Log {
	return &Log{
		config:     cfg,
		perService: map[string][]LogRecord{},
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
	record := LogRecord{
		Service: service,
		Message: msg,
		Time:    time.Now().Unix(),
		ID:      l.curID,
	}
	l.curID++
	l.allRecords = trimLog(append(l.allRecords, record), ls)
	l.perService[service] = trimLog(append(l.perService[service], record), ls)
	l.logLock.Unlock()
	return record.ID, nil
}

// Overview returns the most recent log record per
// service, sorter from least to most recent.
func (l *Log) Overview() []LogRecord {
	l.logLock.RLock()
	var entries []LogRecord
	for _, v := range l.perService {
		entries = append(entries, v[len(v)-1])
	}
	l.logLock.RUnlock()
	sort.Sort(logIDSorter(entries))
	return entries
}

// FullLog returns all of the log records, sorted from
// least to most recent.
func (l *Log) FullLog() []LogRecord {
	l.logLock.RLock()
	res := append([]LogRecord{}, l.allRecords...)
	l.logLock.RUnlock()
	return res
}

// ServiceLog returns the log records for a particular
// service, sorted from least to most recent.
// It fails if there are no log records for the service.
func (l *Log) ServiceLog(name string) ([]LogRecord, error) {
	l.logLock.RLock()
	defer l.logLock.RUnlock()
	entries, ok := l.perService[name]
	if !ok {
		return nil, errors.New("unknown service: " + name)
	}
	return append([]LogRecord{}, entries...), nil
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

func trimLog(log []LogRecord, maxSize int) []LogRecord {
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

type logIDSorter []LogRecord

func (l logIDSorter) Len() int {
	return len(l)
}

func (l logIDSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l logIDSorter) Less(i, j int) bool {
	return l[i].ID < l[j].ID
}
