package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

type TaskProvider interface {
	Lease(owner string, now time.Time, limit int) []protocol.ScheduleItem
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
	if limit < 1 {
		return 0, errors.New("limit must be positive")
	}
	if r.Owner == "" {
		return 0, errors.New("owner is required")
	}
	items := r.Scheduler.Lease(r.Owner, r.Now(), limit)
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := r.Scheduler.Complete(item.TaskID, r.Owner); err != nil {
			_ = r.Scheduler.Fail(item.TaskID, r.Owner, r.Now())
			return 0, err
		}
	}
	return len(items), nil
}

func (r *Runner) String() string {
	return fmt.Sprintf("runner %s", r.Owner)
}
