package domain

import (
	"testing"
	"time"
)

func TestRetryingCanTransitionToCollecting(t *testing.T) {
	task := Task{State: Retrying}
	if err := task.Transition(Collecting, time.Now()); err != nil {
		t.Fatalf("retrying should be allowed to collect again: %v", err)
	}
}

func TestIsActiveIncludesRetrying(t *testing.T) {
	if !IsActive(Retrying) {
		t.Fatal("retrying should be considered active")
	}
}

func TestReplayLatestActiveFindsRetrying(t *testing.T) {
	log := NewReplayLog(10)
	if err := log.Append(ReplayEntry{TaskID: "task-a", Phase: "retrying", Action: "retry", At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	got, ok := log.LatestActive("task-a")
	if !ok || got != "retrying" {
		t.Fatalf("expected retrying active state, got %q ok=%v", got, ok)
	}
}
