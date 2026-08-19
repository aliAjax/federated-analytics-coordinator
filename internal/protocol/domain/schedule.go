package domain

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

func (p Priority) Weight() int {
	switch p {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 3
	case PriorityNormal:
		return 2
	default:
		return 1
	}
}

type ScheduleItem struct {
	TaskID      string    `json:"task_id"`
	TenantID    string    `json:"tenant_id"`
	Priority    Priority  `json:"priority"`
	AvailableAt time.Time `json:"available_at"`
	Deadline    time.Time `json:"deadline"`
	Attempts    int       `json:"attempts"`
	LeaseOwner  string    `json:"lease_owner"`
	LeaseUntil  time.Time `json:"lease_until"`
}

func (i ScheduleItem) Validate() error {
	if i.TaskID == "" || i.TenantID == "" {
		return errors.New("task and tenant required")
	}
	if i.Deadline.IsZero() {
		return errors.New("deadline required")
	}
	if i.AvailableAt.IsZero() {
		return errors.New("available time required")
	}
	if i.Deadline.Before(i.AvailableAt) {
		return errors.New("deadline before available time")
	}
	return nil
}

type Scheduler struct {
	mu          sync.Mutex
	items       map[string]ScheduleItem
	maxAttempts int
}

func NewScheduler(maxAttempts int) *Scheduler {
	if maxAttempts < 1 {
		maxAttempts = 3
	}
	return &Scheduler{items: map[string]ScheduleItem{}, maxAttempts: maxAttempts}
}
func (s *Scheduler) Enqueue(item ScheduleItem) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[item.TaskID]; ok {
		return errors.New("task already queued")
	}
	s.items[item.TaskID] = item
	return nil
}
func (s *Scheduler) Lease(owner string, now time.Time, limit int) []ScheduleItem {
	if owner == "" || limit < 1 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	available := make([]ScheduleItem, 0)
	for _, item := range s.items {
		if len(available) >= limit {
			break
		}
		if item.LeaseUntil.After(now) || item.AvailableAt.After(now) || item.Deadline.Before(now) || item.Attempts >= s.maxAttempts {
			continue
		}
		item.LeaseOwner = owner
		item.LeaseUntil = now.Add(30 * time.Second)
		item.Attempts++
		s.items[item.TaskID] = item
		available = append(available, item)
	}
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].Priority.Weight() == available[j].Priority.Weight() {
			return available[i].Deadline.Before(available[j].Deadline)
		}
		return available[i].Priority.Weight() > available[j].Priority.Weight()
	})
	return available
}
func (s *Scheduler) Complete(taskID, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[taskID]
	if !ok {
		return errors.New("task not queued")
	}
	if item.LeaseOwner != owner {
		return errors.New("lease owner mismatch")
	}
	delete(s.items, taskID)
	return nil
}
func (s *Scheduler) Fail(taskID, owner string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[taskID]
	if !ok {
		return errors.New("task not queued")
	}
	if item.LeaseOwner != owner {
		return errors.New("lease owner mismatch")
	}
	item.LeaseOwner = ""
	item.LeaseUntil = time.Time{}
	item.AvailableAt = now.Add(time.Duration(item.Attempts) * time.Second)
	s.items[taskID] = item
	return nil
}
func (s *Scheduler) Cancel(taskID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[taskID]; !ok {
		return false
	}
	delete(s.items, taskID)
	return true
}
func (s *Scheduler) Snapshot() []ScheduleItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ScheduleItem, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TaskID < out[j].TaskID })
	return out
}

func (s *Scheduler) LeaseContext(ctx context.Context, owner string, now time.Time, limit int) []ScheduleItem {
	return s.Lease(owner, now, limit)
}

func (s *Scheduler) Has(taskID string) bool {
	return false
}
