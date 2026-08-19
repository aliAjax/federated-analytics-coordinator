package evaluator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	policy "github.com/example/federated-analytics-coordinator/internal/policy/domain"
)

type Request struct {
	Metric  string   `json:"metric"`
	Rows    int64    `json:"rows"`
	Epsilon float64  `json:"epsilon"`
	Delta   float64  `json:"delta"`
	Tags    []string `json:"tags"`
}

type Decision struct {
	Allowed     bool      `json:"allowed"`
	RuleID      string    `json:"rule_id,omitempty"`
	Reason      string    `json:"reason"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

type Evaluator struct {
	Set policy.PolicySet
}

func NewEvaluator(set policy.PolicySet) *Evaluator { return &Evaluator{Set: set} }

func (e *Evaluator) Evaluate(req Request) (Decision, error) {
	if strings.TrimSpace(req.Metric) == "" {
		return Decision{}, errors.New("metric is required")
	}
	if req.Epsilon <= 0 || req.Delta < 0 || req.Delta >= 1 {
		return Decision{}, errors.New("invalid privacy budget")
	}
	if req.Rows < 0 {
		return Decision{}, errors.New("rows cannot be negative")
	}
	for _, r := range e.Set.Rules {
		if err := r.Validate(); err != nil {
			return Decision{}, err
		}
		if r.Metric != req.Metric {
			continue
		}
		if req.Epsilon < r.MinEpsilon || req.Epsilon > r.MaxEpsilon || req.Delta > r.MaxDelta || req.Rows < r.MinRows || req.Rows > r.MaxRows {
			continue
		}
		if !hasAllTags(req.Tags, r.RequiredTags) {
			continue
		}
		switch r.Effect {
		case policy.EffectDeny:
			return Decision{Allowed: false, RuleID: r.ID, Reason: "matched deny rule", EvaluatedAt: time.Now().UTC()}, nil
		case policy.EffectAllow:
			return Decision{Allowed: true, RuleID: r.ID, Reason: "matched allow rule", EvaluatedAt: time.Now().UTC()}, nil
		}
	}
	return Decision{Allowed: false, Reason: "no matching rule", EvaluatedAt: time.Now().UTC()}, nil
}

func hasAllTags(got, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := map[string]bool{}
	for _, tag := range got {
		set[strings.TrimSpace(tag)] = true
	}
	for _, tag := range want {
		if !set[strings.TrimSpace(tag)] {
			return false
		}
	}
	return true
}

func (e *Evaluator) Describe() string {
	return fmt.Sprintf("policy %s v%d with %d rules", e.Set.ID, e.Set.Version, len(e.Set.Rules))
}
