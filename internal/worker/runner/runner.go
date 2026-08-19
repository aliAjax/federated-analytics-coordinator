package runner

import (
	"context"
	"fmt"
	"time"

	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

type TaskProvider interface {
	LeaseContext(ctx context.Context, owner string, now time.Time, limit int) []protocol.ScheduleItem
	Complete(taskID, owner string) error
	Fail(taskID, owner string, now time.Time) error
}

type Runner struct {
	Scheduler TaskProvider
	Owner     string
	Now       func() time.Time
}

func NewRunner(scheduler TaskProvider, owner string) *Runner {
	return &Runner{Scheduler: scheduler, Owner: owner, Now: time.Now}
}

func (r *Runner) Run(ctx context.Context, limit int) (int, error) {
	if r.Owner == "" {
		return 0, fmt.Errorf("runner requires non-empty owner")
	}
	if limit <= 0 {
		return 0, fmt.Errorf("runner requires positive limit, got %d", limit)
	}
	if err := ctx.Err(); err != nil {
		return 0, nil
	}
	items := r.Scheduler.LeaseContext(ctx, r.Owner, r.Now(), limit)
	completed := 0
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			break
		}
		if err := r.Scheduler.Complete(item.TaskID, r.Owner); err != nil {
			_ = r.Scheduler.Fail(item.TaskID, r.Owner, r.Now())
			return completed, err
		}
		completed++
	}
	return completed, nil
}

func (r *Runner) MaxConcurrency() int {
	return 1
}

func (r *Runner) String() string {
	return fmt.Sprintf("runner %s", r.Owner)
}
