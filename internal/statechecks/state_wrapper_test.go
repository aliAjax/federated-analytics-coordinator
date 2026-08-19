package statechecks

import (
	"testing"
	"time"

	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func TestRetryingCanTransitionToCollecting(t *testing.T) {
	task := protocol.Task{State: protocol.Retrying}
	if err := task.Transition(protocol.Collecting, time.Now()); err != nil {
		t.Fatalf("retrying should be allowed to collect again: %v", err)
	}
}

func TestIsActiveIncludesRetrying(t *testing.T) {
	if !protocol.IsActive(protocol.Retrying) {
		t.Fatal("retrying should be considered active")
	}
}

func TestReplayLatestActiveFindsRetrying(t *testing.T) {
	log := protocol.NewReplayLog(10)
	if err := log.Append(protocol.ReplayEntry{TaskID: "task-a", Phase: "retrying", Action: "retry", At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	got, ok := log.LatestActive("task-a")
	if !ok || got != "retrying" {
		t.Fatalf("expected retrying active state, got %q ok=%v", got, ok)
	}
}
