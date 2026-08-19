package applicationtest

import (
	"fmt"
	"testing"
	"time"

	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
	application "github.com/example/federated-analytics-coordinator/internal/protocol/application"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

type repo struct{ tasks map[string]protocol.Task }

func (m *repo) CreateParticipant(federation.Participant) error { return nil }
func (m *repo) FindParticipant(string, string) (federation.Participant, error) {
	return federation.Participant{}, nil
}
func (m *repo) CreateTask(t protocol.Task) error { m.tasks[t.ID] = t; return nil }
func (m *repo) FindTask(_, id string) (protocol.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return protocol.Task{}, fmt.Errorf("not found")
	}
	return t, nil
}
func (m *repo) SaveTask(t protocol.Task) error { m.tasks[t.ID] = t; return nil }

func TestCompleteTaskDoesNotSwallowBudgetError(t *testing.T) {
	r := &repo{tasks: map[string]protocol.Task{}}
	task := protocol.Task{ID: "task-a", TenantID: "tenant", TemplateID: "template", Metric: protocol.Count, State: protocol.Revealing, Participants: []string{"p1", "p2"}, MinimumParticipants: 2, Deadline: time.Now().Add(time.Hour), Budget: protocol.Budget{Epsilon: 1, Spent: 2}}
	_ = r.CreateTask(task)
	s := application.New(r, func() string { return "id" })
	if err := s.CompleteTask("tenant", "task-a"); err == nil {
		t.Fatal("expected budget overspent error")
	}
	got, _ := r.FindTask("tenant", "task-a")
	if got.State != protocol.Revealing {
		t.Fatalf("task should remain revealing, got %s", got.State)
	}
}

func TestCanCompleteRejectsOverspent(t *testing.T) {
	s := application.New(nil, func() string { return "id" })
	task := protocol.Task{State: protocol.Revealing, Budget: protocol.Budget{Epsilon: 1, Spent: 2}}
	if s.CanComplete(task) {
		t.Fatal("expected overspent task to be blocked")
	}
}

func TestCompletionReasonBudgetOverspent(t *testing.T) {
	s := application.New(nil, func() string { return "id" })
	task := protocol.Task{State: protocol.Revealing, Budget: protocol.Budget{Epsilon: 1, Spent: 2}}
	if got := s.CompletionReason(task); got != "budget overspent" {
		t.Fatalf("got %q", got)
	}
}

func TestCompletionAllowedRejectsNegativeDelta(t *testing.T) {
	s := application.New(nil, func() string { return "id" })
	if err := s.CompletionAllowed(protocol.Budget{Epsilon: 1, Delta: -1}); err == nil {
		t.Fatal("expected negative delta error")
	}
}
