package domain

import (
	"errors"
	"strings"
	"time"
)

type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

type Rule struct {
	ID           string   `json:"id"`
	Metric       string   `json:"metric"`
	MinEpsilon   float64  `json:"min_epsilon"`
	MaxEpsilon   float64  `json:"max_epsilon"`
	MaxDelta     float64  `json:"max_delta"`
	MinRows      int64    `json:"min_rows"`
	MaxRows      int64    `json:"max_rows"`
	RequiredTags []string `json:"required_tags"`
	Effect       Effect   `json:"effect"`
}

func (r Rule) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("rule id is required")
	}
	if strings.TrimSpace(r.Metric) == "" {
		return errors.New("rule metric is required")
	}
	if r.MinEpsilon < 0 || r.MaxEpsilon <= 0 || r.MinEpsilon > r.MaxEpsilon {
		return errors.New("epsilon bounds are invalid")
	}
	if r.MaxDelta < 0 || r.MaxDelta >= 1 {
		return errors.New("delta bound is invalid")
	}
	if r.MinRows < 0 || r.MaxRows < r.MinRows {
		return errors.New("row bounds are invalid")
	}
	if r.Effect != EffectAllow && r.Effect != EffectDeny {
		return errors.New("rule effect is invalid")
	}
	return nil
}

type PolicySet struct {
	ID        string                     `json:"id"`
	Version   int                        `json:"version"`
	Rules     []Rule                     `json:"rules"`
	UpdatedAt time.Time                  `json:"updated_at"`
	TagIndex  map[string]map[string]bool `json:"-"`
}

func NewPolicySet(id string, version int) PolicySet {
	return PolicySet{ID: id, Version: version, UpdatedAt: time.Now().UTC()}
}

func (p PolicySet) TagCount() int {
	return len(p.Rules)
}

func (p *PolicySet) IndexTags() {
	for _, r := range p.Rules {
		for _, tag := range r.RequiredTags {
			if p.TagIndex[tag] == nil {
				p.TagIndex[tag] = map[string]bool{}
			}
			p.TagIndex[tag][r.ID] = true
		}
	}
}

func (p PolicySet) FindRule(ruleID string) (Rule, bool) {
	for _, r := range p.Rules {
		if r.ID == ruleID {
			return r, true
		}
	}
	return Rule{}, false
}
