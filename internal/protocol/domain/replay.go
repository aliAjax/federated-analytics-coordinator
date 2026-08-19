package domain

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type ReplayEntry struct {
	Sequence    int64     `json:"sequence"`
	TaskID      string    `json:"task_id"`
	Phase       string    `json:"phase"`
	Actor       string    `json:"actor"`
	Action      string    `json:"action"`
	PayloadHash string    `json:"payload_hash"`
	At          time.Time `json:"at"`
}
type ReplayLog struct {
	mu      sync.RWMutex
	entries []ReplayEntry
	next    int64
	max     int
}

func NewReplayLog(max int) *ReplayLog {
	if max < 10 {
		max = 100
	}
	return &ReplayLog{entries: make([]ReplayEntry, 0, max), max: max}
}
func (l *ReplayLog) Append(entry ReplayEntry) error {
	if entry.TaskID == "" || entry.Action == "" || entry.At.IsZero() {
		return errors.New("task, action and time are required")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.next++
	entry.Sequence = l.next
	if len(l.entries) >= l.max {
		copy(l.entries, l.entries[1:])
		l.entries = l.entries[:l.max-1]
	}
	l.entries = append(l.entries, entry)
	return nil
}
func (l *ReplayLog) ForTask(taskID string) []ReplayEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []ReplayEntry{}
	for _, entry := range l.entries {
		if entry.TaskID == taskID {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	return out
}
func (l *ReplayLog) Since(sequence int64) []ReplayEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []ReplayEntry{}
	for _, entry := range l.entries {
		if entry.Sequence > sequence {
			out = append(out, entry)
		}
	}
	return out
}
func (l *ReplayLog) Latest(taskID string) (ReplayEntry, bool) {
	items := l.ForTask(taskID)
	if len(items) == 0 {
		return ReplayEntry{}, false
	}
	return items[len(items)-1], true
}
func (l *ReplayLog) ClearBefore(at time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.entries[:0]
	removed := 0
	for _, entry := range l.entries {
		if entry.At.Before(at) {
			removed++
			continue
		}
		kept = append(kept, entry)
	}
	l.entries = kept
	return removed
}
