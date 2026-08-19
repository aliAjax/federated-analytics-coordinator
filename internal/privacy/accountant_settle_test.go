package privacy

import (
	"errors"
	"testing"
	"time"
)

func TestLedgerNotFoundWrapsSentinel(t *testing.T) {
	l := NewLedger()
	err := l.Commit("missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected wrapped ErrEntryNotFound, got %v", err)
	}
}

func TestSettleDoesNotCommitInvalidEntry(t *testing.T) {
	a := NewAccountant(10, 2)
	l := NewLedger()
	entry := LedgerEntry{ID: "entry-a", TaskID: "task-a", Participant: "p1", Epsilon: 1, Delta: 1, At: time.Now().UTC()}
	if err := a.Settle(l, entry); err == nil {
		t.Fatal("expected invalid delta error")
	}
	items := l.ForTask("task-a")
	if len(items) != 1 {
		t.Fatalf("expected one reserved ledger item, got %d", len(items))
	}
	if items[0].Committed {
		t.Fatal("invalid entry must not be committed")
	}
}
