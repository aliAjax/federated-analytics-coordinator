package domain

import (
	"fmt"
	"time"
)

type State string

const (
	Draft       State = "draft"
	Collecting  State = "collecting"
	Masked      State = "masked"
	Aggregating State = "aggregating"
	Revealing   State = "revealing"
	Completed   State = "completed"
	Aborted     State = "aborted"
)

type Metric string

const (
	Count     Metric = "count"
	Mean      Metric = "mean"
	Histogram Metric = "histogram"
	Quantile  Metric = "quantile"
	Gradient  Metric = "gradient"
)

type Budget struct {
	Epsilon  float64 `json:"epsilon"`
	Delta    float64 `json:"delta"`
	Reserved float64 `json:"reserved"`
	Spent    float64 `json:"spent"`
}

func (b Budget) Available() float64 { return b.Epsilon - b.Reserved - b.Spent }
func (b *Budget) Reserve(v float64) error {
	if v <= 0 {
		return fmt.Errorf("reservation must be positive")
	}
	if b.Available() < v {
		return fmt.Errorf("privacy budget exhausted")
	}
	b.Reserved += v
	return nil
}
func (b *Budget) Commit(v float64) error {
	if v <= 0 || b.Reserved < v {
		return fmt.Errorf("invalid budget commit")
	}
	b.Reserved -= v
	b.Spent += v
	return nil
}
func (b *Budget) Release(v float64) {
	if v > 0 && b.Reserved >= v {
		b.Reserved -= v
	}
}

type Task struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenant_id"`
	TemplateID          string    `json:"template_id"`
	Metric              Metric    `json:"metric"`
	State               State     `json:"state"`
	Participants        []string  `json:"participants"`
	MinimumParticipants int       `json:"minimum_participants"`
	Deadline            time.Time `json:"deadline"`
	Budget              Budget    `json:"budget"`
	Revision            int64     `json:"revision"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	AbortReason         string    `json:"abort_reason,omitempty"`
}

func (t *Task) Transition(next State, now time.Time) error {
	allowed := map[State][]State{Draft: {Collecting, Aborted}, Collecting: {Masked, Aborted}, Masked: {Aggregating, Aborted}, Aggregating: {Revealing, Aborted}, Revealing: {Completed, Aborted}}
	for _, v := range allowed[t.State] {
		if v == next {
			t.State = next
			t.UpdatedAt = now
			t.Revision++
			return nil
		}
	}
	return fmt.Errorf("invalid state transition from %s to %s", t.State, next)
}
func (t Task) Validate() error {
	if t.ID == "" || t.TenantID == "" || t.TemplateID == "" {
		return fmt.Errorf("task identity is incomplete")
	}
	if t.MinimumParticipants < 2 {
		return fmt.Errorf("minimum participants must be at least two")
	}
	if len(t.Participants) < t.MinimumParticipants {
		return fmt.Errorf("insufficient task participants")
	}
	if t.Deadline.IsZero() {
		return fmt.Errorf("deadline required")
	}
	return nil
}
