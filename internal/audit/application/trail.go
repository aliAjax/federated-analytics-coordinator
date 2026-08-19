package engine

import (
	"errors"
	"strings"
	"time"

	audit "github.com/example/federated-analytics-coordinator/internal/audit/domain"
)

// Trail is an append-only, hash-chained audit log.
type Trail struct {
	entries []audit.Entry
	last    string
}

func NewTrail() *Trail { return &Trail{} }

func (t *Trail) Append(entry audit.Entry) (audit.Entry, error) {
	if strings.TrimSpace(entry.TenantID) == "" {
		return audit.Entry{}, errors.New("tenant id is required")
	}
	if strings.TrimSpace(entry.Action) == "" {
		return audit.Entry{}, errors.New("action is required")
	}
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	entry.PreviousHash = t.last
	entry.Seal()
	if entry.Hash == "" {
		return audit.Entry{}, errors.New("entry sealing failed")
	}
	t.entries = append(t.entries, entry)
	t.last = entry.Hash
	return entry, nil
}

func (t *Trail) Verify() error {
	prev := ""
	for _, entry := range t.entries {
		want := entry.PreviousHash
		if want != prev {
			return errors.New("audit chain is broken")
		}
		probe := entry
		probe.Hash = ""
		probe.Seal()
		if probe.Hash != entry.Hash {
			return errors.New("audit entry hash mismatch")
		}
		prev = entry.Hash
	}
	return nil
}

func (t *Trail) List(tenant string) []audit.Entry {
	out := []audit.Entry{}
	for _, entry := range t.entries {
		if tenant == "" || entry.TenantID == tenant {
			out = append(out, entry)
		}
	}
	return out
}

func (t *Trail) Len() int { return len(t.entries) }
