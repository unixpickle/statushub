package main

import (
	"errors"
	"sync"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

type MediaRecord struct {
	statushub.MediaRecord
	Data []byte `json:"-"`
}

// Log maintains a history of statushub.LogRecords.
type Log struct {
	config     *Config
	logLock    sync.RWMutex
	curID      int
	perService map[string][]statushub.LogRecord
	allRecords []statushub.LogRecord
	media      map[string][]MediaRecord

	serviceChans map[string]chan struct{}
	globalChan   chan struct{}
}

// NewLog creates a log which depends on a configuration
// to get the maximum log size.
func NewLog(cfg *Config) *Log {
	return &Log{
		config:       cfg,
		perService:   map[string][]statushub.LogRecord{},
		media:        map[string][]MediaRecord{},
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

// AddMedia adds a media record.
func (l *Log) AddMedia(folder, filename, mime string, data []byte, replace bool) (int, error) {
	cacheSize := l.config.MediaCache()

	// See comment in Add().

	l.logLock.Lock()
	record := MediaRecord{
		MediaRecord: statushub.MediaRecord{
			Folder:   folder,
			Filename: filename,
			Mime:     mime,
			Time:     time.Now().Unix(),
			ID:       l.curID,
		},
		Data: data,
	}
	l.curID++
	media := l.media[folder]
	if replace {
		media = removeMedia(media, filename)
	}
	media = append(media, record)
	media = trimMedia(media, cacheSize)
	l.media[folder] = media
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

// DeleteMedia deletes a media entry.
// It fails if the folder does not exist.
func (l *Log) DeleteMedia(folder string) error {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	if _, ok := l.media[folder]; !ok {
		return errors.New("no such media folder: " + folder)
	}
	delete(l.media, folder)
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
	essentials.VoodooSort(entries, func(i, j int) bool {
		return entries[i].ID > entries[j].ID
	})
	return entries
}

// MediaOverview returns the most recent media record per
// folder, sorted from most to least recent.
func (l *Log) MediaOverview() []statushub.MediaRecord {
	l.logLock.RLock()
	var entries []statushub.MediaRecord
	for _, v := range l.media {
		entries = append(entries, v[len(v)-1].MediaRecord)
	}
	l.logLock.RUnlock()
	essentials.VoodooSort(entries, func(i, j int) bool {
		return entries[i].ID > entries[j].ID
	})
	return entries
}

// FullLog returns all of the log records, sorted from
// most to least recent.
func (l *Log) FullLog() []statushub.LogRecord {
	l.logLock.RLock()
	res := append([]statushub.LogRecord{}, l.allRecords...)
	l.logLock.RUnlock()
	essentials.Reverse(res)
	return res
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
	entries = append([]statushub.LogRecord{}, entries...)
	essentials.Reverse(entries)
	return entries, nil
}

// MediaLog returns the media records for a folder.
// It fails if there are no media records for the folder.
func (l *Log) MediaLog(folder string) ([]statushub.MediaRecord, error) {
	l.logLock.RLock()
	defer l.logLock.RUnlock()
	entries, ok := l.media[folder]
	if !ok {
		return nil, errors.New("unknown media folder: " + folder)
	}
	res := make([]statushub.MediaRecord, len(entries))
	for i, x := range entries {
		res[len(entries)-(i+1)] = x.MediaRecord
	}
	return res, nil
}

// MediaRecord looks up the media record by ID.
func (l *Log) MediaRecord(id int) *MediaRecord {
	l.logLock.RLock()
	defer l.logLock.RUnlock()
	for _, records := range l.media {
		for _, record := range records {
			if record.ID == id {
				return &record
			}
		}
	}
	return nil
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

// MediaCacheUpdated directs the log to delete media
// records as needed to accommodate the new cache size.
func (l *Log) MediaCacheUpdated() {
	cacheSize := l.config.MediaCache()
	l.logLock.Lock()
	for k, v := range l.media {
		l.media[k] = trimMedia(v, cacheSize)
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

func trimMedia(log []MediaRecord, cacheSize int) []MediaRecord {
	if cacheSize == 0 {
		return log
	}
	for len(log) > 1 {
		totalSize := 0
		for _, item := range log {
			totalSize += len(item.Data)
		}
		if totalSize > cacheSize {
			essentials.OrderedDelete(&log, 0)
		} else {
			break
		}
	}
	return log
}

func removeMedia(log []MediaRecord, filename string) []MediaRecord {
	for i := 0; i < len(log); i++ {
		if log[i].Filename == filename {
			essentials.OrderedDelete(&log, i)
			i--
		}
	}
	return log
}
