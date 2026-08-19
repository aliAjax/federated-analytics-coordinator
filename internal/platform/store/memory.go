package store

import (
	"fmt"
	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
	"sync"
)

type Memory struct {
	mu           sync.RWMutex
	participants map[string]federation.Participant
	tasks        map[string]protocol.Task
}

func New() *Memory {
	return &Memory{participants: map[string]federation.Participant{}, tasks: map[string]protocol.Task{}}
}
func k(tenant, id string) string { return tenant + "/" + id }
func (m *Memory) CreateParticipant(p federation.Participant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := k(p.OrganizationID, p.ID)
	if _, ok := m.participants[key]; ok {
		return fmt.Errorf("participant exists")
	}
	m.participants[key] = p
	return nil
}
func (m *Memory) FindParticipant(tenant, id string) (federation.Participant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.participants[k(tenant, id)]
	if !ok {
		return federation.Participant{}, fmt.Errorf("participant not found")
	}
	return p, nil
}
func (m *Memory) CreateTask(t protocol.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := k(t.TenantID, t.ID)
	if _, ok := m.tasks[key]; ok {
		return fmt.Errorf("task exists")
	}
	m.tasks[key] = t
	return nil
}
func (m *Memory) FindTask(tenant, id string) (protocol.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.tasks[k(tenant, id)]
	if !ok {
		return protocol.Task{}, protocol.ErrNotFound
	}
	return cloneTask(v), nil
}
func (m *Memory) SaveTask(t protocol.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := k(t.TenantID, t.ID)
	if _, ok := m.tasks[key]; !ok {
		return fmt.Errorf("task not found")
	}
	m.tasks[key] = cloneTask(t)
	return nil
}
func cloneTask(t protocol.Task) protocol.Task {
	t.Participants = append([]string(nil), t.Participants...)
	return t
}
