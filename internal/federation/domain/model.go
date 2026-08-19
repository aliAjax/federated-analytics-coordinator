package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ParticipantStatus string

const (
	ParticipantActive    ParticipantStatus = "active"
	ParticipantRevoked   ParticipantStatus = "revoked"
	ParticipantSuspended ParticipantStatus = "suspended"
)

type Capability struct {
	Metric                  string `json:"metric"`
	MaxRows                 int64  `json:"max_rows"`
	SupportsDropoutRecovery bool   `json:"supports_dropout_recovery"`
	Version                 string `json:"version"`
}
type Participant struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	Name           string            `json:"name"`
	PublicKey      []byte            `json:"public_key"`
	Status         ParticipantStatus `json:"status"`
	Capabilities   []Capability      `json:"capabilities"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Revision       int64             `json:"revision"`
}

func (p Participant) Validate() error {
	if p.ID == "" || p.OrganizationID == "" || strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("participant id, organization and name are required")
	}
	if len(p.PublicKey) != 32 {
		return fmt.Errorf("Ed25519 public key must have 32 bytes")
	}
	if p.Status == "" {
		return fmt.Errorf("participant status is required")
	}
	seen := map[string]bool{}
	for _, c := range p.Capabilities {
		if c.Metric == "" {
			return fmt.Errorf("capability metric required")
		}
		if seen[c.Metric] {
			return fmt.Errorf("duplicate capability %s", c.Metric)
		}
		seen[c.Metric] = true
	}
	return nil
}
func (p Participant) Supports(metric string) bool {
	for _, c := range p.Capabilities {
		if c.Metric == metric {
			return true
		}
	}
	return false
}
func NormalizeCapabilities(v []Capability) []Capability {
	out := append([]Capability(nil), v...)
	sort.Slice(out, func(i, j int) bool { return out[i].Metric < out[j].Metric })
	return out
}
