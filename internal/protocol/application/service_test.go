package application

import (
	"fmt"
	"testing"
	"time"

	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func TestCompleteTaskDoesNotSwallowBudgetError(t *testing.T) {
	repo := newMemoryRepo()
	task := protocol.Task{
		ID: "task-a", TenantID: "tenant", TemplateID: "template", Metric: protocol.Count,
		State: protocol.Revealing, Participants: []string{"p1", "p2"}, MinimumParticipants: 2,
		Deadline: time.Now().Add(time.Hour), Budget: protocol.Budget{Epsilon: 1, Spent: 2},
	}
	if err := repo.CreateTask(task); err != nil {
		t.Fatal(err)
	}
	s := New(repo, func() string { return "id" })
	if err := s.CompleteTask("tenant", "task-a"); err == nil {
		t.Fatal("expected budget overspent error")
	}
	got, _ := repo.FindTask("tenant", "task-a")
	if got.State != protocol.Revealing {
		t.Fatalf("task should remain revealing on error, got %s", got.State)
	}
}

type memoryRepo struct {
	tasks map[string]protocol.Task
}

func newMemoryRepo() *memoryRepo                                       { return &memoryRepo{tasks: map[string]protocol.Task{}} }
func (m *memoryRepo) CreateParticipant(p federation.Participant) error { return nil }
func (m *memoryRepo) FindParticipant(string, string) (federation.Participant, error) {
	return federation.Participant{}, nil
}
func (m *memoryRepo) CreateTask(t protocol.Task) error { m.tasks[t.ID] = t; return nil }
func (m *memoryRepo) FindTask(_, id string) (protocol.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return protocol.Task{}, fmt.Errorf("not found")
	}
	return t, nil
}
func (m *memoryRepo) SaveTask(t protocol.Task) error { m.tasks[t.ID] = t; return nil }
