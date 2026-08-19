package privacy

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type LedgerEntry struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	Participant string    `json:"participant"`
	Epsilon     float64   `json:"epsilon"`
	Delta       float64   `json:"delta"`
	At          time.Time `json:"at"`
	Committed   bool      `json:"committed"`
}
type Ledger struct {
	mu      sync.RWMutex
	entries map[string]LedgerEntry
}

func NewLedger() *Ledger { return &Ledger{entries: map[string]LedgerEntry{}} }
func (l *Ledger) Reserve(e LedgerEntry) error {
	if e.ID == "" || e.TaskID == "" || e.Participant == "" || e.Epsilon <= 0 || e.At.IsZero() {
		return errors.New("invalid ledger entry")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.entries[e.ID]; ok {
		return errors.New("ledger entry already exists")
	}
	l.entries[e.ID] = e
	return nil
}
func (l *Ledger) Commit(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[id]
	if !ok {
		return errors.New("ledger entry not found")
	}
	e.Committed = true
	l.entries[id] = e
	return nil
}
func (l *Ledger) Release(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.entries[id]; !ok {
		return errors.New("ledger entry not found")
	}
	delete(l.entries, id)
	return nil
}
func (l *Ledger) ForTask(task string) []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := []LedgerEntry{}
	for _, e := range l.entries {
		if e.TaskID == task {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}
func (l *Ledger) Totals(task string) (float64, float64) {
	var epsilon, delta float64
	for _, e := range l.ForTask(task) {
		if e.Committed {
			epsilon += e.Epsilon
			delta += e.Delta
		}
	}
	return epsilon, delta
}

func (l *Ledger) Count(task string) int { return len(l.ForTask(task)) }
func (l *Ledger) Clear(task string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for id, e := range l.entries {
		if e.TaskID == task {
			delete(l.entries, id)
			n++
		}
	}
	return n
}
