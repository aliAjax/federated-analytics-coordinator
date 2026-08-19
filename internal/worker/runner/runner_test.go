package runner

import (
	"context"
	"testing"
	"time"

	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func TestRunnerHonorsCanceledContext(t *testing.T) {
	sched := protocol.NewScheduler(3)
	now := time.Now()
	for _, id := range []string{"task-a", "task-b"} {
		if err := sched.Enqueue(protocol.ScheduleItem{TaskID: id, TenantID: "tenant", Priority: protocol.PriorityNormal, AvailableAt: now.Add(-time.Minute), Deadline: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := NewRunner(sched, "worker").Run(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("canceled context still leased %d tasks", got)
	}
}

func TestRunnerRejectsInvalidLimit(t *testing.T) {
	sched := protocol.NewScheduler(3)
	_, err := NewRunner(sched, "worker").Run(context.Background(), 0)
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestRunnerRequiresOwner(t *testing.T) {
	sched := protocol.NewScheduler(3)
	_, err := NewRunner(sched, "").Run(context.Background(), 5)
	if err == nil {
		t.Fatal("expected owner error")
	}
}

func TestSchedulerHasTask(t *testing.T) {
	sched := protocol.NewScheduler(3)
	now := time.Now()
	if err := sched.Enqueue(protocol.ScheduleItem{TaskID: "task-a", TenantID: "tenant", Priority: protocol.PriorityNormal, AvailableAt: now, Deadline: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if !sched.Has("task-a") {
		t.Fatal("expected scheduler to have task-a")
	}
}

func TestRunnerMaxConcurrency(t *testing.T) {
	if got := NewRunner(protocol.NewScheduler(3), "worker").MaxConcurrency(); got != 1 {
		t.Fatalf("expected max concurrency 1, got %d", got)
	}
}
