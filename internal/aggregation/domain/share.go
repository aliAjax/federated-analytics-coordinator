package domain

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Share struct {
	ID            string    `json:"id"`
	TaskID        string    `json:"task_id"`
	ParticipantID string    `json:"participant_id"`
	Round         int       `json:"round"`
	Value         int64     `json:"value"`
	Count         int64     `json:"count"`
	Nonce         string    `json:"nonce"`
	Signature     []byte    `json:"signature"`
	ReceivedAt    time.Time `json:"received_at"`
}

func (s Share) Canonical() []byte {
	return []byte(fmt.Sprintf("%s|%s|%d|%d|%d|%s", s.TaskID, s.ParticipantID, s.Round, s.Value, s.Count, s.Nonce))
}
func (s Share) Validate() error {
	if s.TaskID == "" || s.ParticipantID == "" || s.Nonce == "" {
		return fmt.Errorf("share identity and nonce required")
	}
	if s.Round < 1 {
		return fmt.Errorf("round must be positive")
	}
	if s.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	return nil
}

type Aggregate struct {
	TaskID       string    `json:"task_id"`
	Value        int64     `json:"value"`
	Count        int64     `json:"count"`
	Participants []string  `json:"participants"`
	Precision    string    `json:"precision"`
	CreatedAt    time.Time `json:"created_at"`
}

func Sum(task string, shares []Share, now time.Time) (Aggregate, error) {
	seen := map[string]bool{}
	var value, count int64
	ids := make([]string, 0, len(shares))
	for _, s := range shares {
		if s.TaskID != task {
			return Aggregate{}, fmt.Errorf("share belongs to another task")
		}
		if seen[s.ParticipantID] {
			return Aggregate{}, fmt.Errorf("duplicate participant share")
		}
		seen[s.ParticipantID] = true
		if (s.Value > 0 && value > math.MaxInt64-s.Value) || (s.Value < 0 && value < math.MinInt64-s.Value) {
			return Aggregate{}, fmt.Errorf("aggregate overflow")
		}
		value += s.Value
		if count > math.MaxInt64-s.Count {
			return Aggregate{}, fmt.Errorf("count overflow")
		}
		count += s.Count
		ids = append(ids, s.ParticipantID)
	}
	sort.Strings(ids)
	return Aggregate{TaskID: task, Value: value, Count: count, Participants: ids, Precision: "int64 exact before differential privacy", CreatedAt: now}, nil
}
