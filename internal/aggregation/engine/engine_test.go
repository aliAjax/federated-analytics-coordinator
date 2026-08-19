package engine

import (
	"reflect"
	"testing"

	aggregate "github.com/example/federated-analytics-coordinator/internal/aggregation/domain"
)

func TestMergeSharesDoesNotMutateInput(t *testing.T) {
	shares := []aggregate.Share{
		{TaskID: "task-a", ParticipantID: "p1", Round: 1, Value: 10, Count: 1, Nonce: "n1"},
		{TaskID: "task-a", Round: 1, Value: 20, Count: 1, Nonce: "n2"},
		{TaskID: "task-a", ParticipantID: "p2", Round: 1, Value: 30, Count: 1, Nonce: "n3"},
	}
	original := append([]aggregate.Share(nil), shares...)
	got, err := NewEngine().MergeShares("task-a", shares)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(shares, original) {
		t.Fatalf("mutated input: %#v", shares)
	}
	if got.Count != 2 {
		t.Fatalf("unexpected count %d", got.Count)
	}
}

func TestMergeSharesRejectsEmptyTaskID(t *testing.T) {
	_, err := NewEngine().MergeShares("", nil)
	if err == nil {
		t.Fatal("expected empty task id error")
	}
}

func TestCloneSharesReturnsIndependentCopy(t *testing.T) {
	shares := []aggregate.Share{{ParticipantID: "p1"}}
	clone := aggregate.CloneShares(shares)
	clone[0].ParticipantID = "changed"
	if shares[0].ParticipantID != "p1" {
		t.Fatal("clone should not share backing array")
	}
}

func TestEngineParticipantCountIgnoresEmptyIDs(t *testing.T) {
	shares := []aggregate.Share{{ParticipantID: "p1"}, {ParticipantID: "p2"}, {}}
	if got := NewEngine().ParticipantCount(shares); got != 2 {
		t.Fatalf("expected 2 participants, got %d", got)
	}
}

func TestUniqueParticipantsRemovesDuplicates(t *testing.T) {
	shares := []aggregate.Share{{ParticipantID: "p1"}, {ParticipantID: "p2"}, {ParticipantID: "p1"}}
	got := aggregate.UniqueParticipants(shares)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique participants, got %v", got)
	}
}
