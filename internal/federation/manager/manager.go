package manager

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
)

// Manager negotiates participant capabilities and quorum plans.
type Manager struct {
	mu           sync.RWMutex
	participants map[string]federation.Participant
}

func NewManager() *Manager {
	return &Manager{participants: map[string]federation.Participant{}}
}

func (m *Manager) AddParticipant(p federation.Participant) error {
	if err := p.Validate(); err != nil {
		return err
	}
	key := strings.TrimSpace(p.ID)
	if key == "" {
		return errors.New("participant id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.participants[key]; ok {
		return fmt.Errorf("participant already exists: %s", key)
	}
	clone := p.Clone()
	clone.Capabilities = federation.NormalizeCapabilities(clone.Capabilities)
	m.participants[key] = clone
	return nil
}

type QuorumPlan struct {
	Metric              string   `json:"metric"`
	Required            int      `json:"required"`
	Eligible            []string `json:"eligible"`
	SupportsDropout     bool     `json:"supports_dropout"`
	MissingCapabilities []string `json:"missing_capabilities,omitempty"`
}

func (m *Manager) PlanQuorum(metric string, required int) (QuorumPlan, error) {
	if strings.TrimSpace(metric) == "" {
		return QuorumPlan{}, errors.New("metric is required")
	}
	if required < 2 {
		return QuorumPlan{}, errors.New("quorum must be at least two")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	plan := QuorumPlan{Metric: metric, Required: required, SupportsDropout: true}
	for _, p := range m.participants {
		if p.Status != federation.ParticipantActive {
			continue
		}
		if p.Supports(metric) {
			plan.Eligible = append(plan.Eligible, p.ID)
			cap := capabilityForMetric(p.Capabilities, metric)
			if !cap.SupportsDropoutRecovery {
				plan.SupportsDropout = false
			}
		} else {
			plan.MissingCapabilities = append(plan.MissingCapabilities, p.ID)
		}
	}
	sort.Strings(plan.Eligible)
	sort.Strings(plan.MissingCapabilities)
	if len(plan.Eligible) < required {
		return plan, fmt.Errorf("quorum not satisfiable: %d eligible, %d required", len(plan.Eligible), required)
	}
	return plan, nil
}

func capabilityForMetric(caps []federation.Capability, metric string) federation.Capability {
	for _, c := range caps {
		if c.Metric == metric {
			return c
		}
	}
	return federation.Capability{}
}

func (m *Manager) SupportedMetrics() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := map[string]bool{}
	for _, p := range m.participants {
		if p.Status != federation.ParticipantActive {
			continue
		}
		for _, c := range p.Capabilities {
			seen[c.Metric] = true
		}
	}
	out := make([]string, 0, len(seen))
	for metric := range seen {
		out = append(out, metric)
	}
	sort.Strings(out)
	return out
}

func (m *Manager) Snapshot() []federation.Participant {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]federation.Participant, 0, len(m.participants))
	for _, p := range m.participants {
		out = append(out, p.Clone())
	}
	return out
}
