package evaluator

import (
	"testing"

	policy "github.com/example/federated-analytics-coordinator/internal/policy/domain"
)

func TestEvaluatorIndexesRequiredTagsWithoutPanic(t *testing.T) {
	set := policy.NewPolicySet("release-rules", 1)
	set.Rules = []policy.Rule{{ID: "allow-prod-count", Metric: "count", MinEpsilon: 0.1, MaxEpsilon: 10, MaxDelta: 0.01, MinRows: 1, MaxRows: 1000, RequiredTags: []string{"prod", "signed"}, Effect: policy.EffectAllow}}
	ev := NewEvaluator(set)
	got, err := ev.Evaluate(Request{Metric: "count", Rows: 10, Epsilon: 1, Delta: 0.001, Tags: []string{"signed", "prod"}})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Allowed || got.RuleID != "allow-prod-count" {
		t.Fatalf("expected allow rule to match, got %+v", got)
	}
}

func TestPolicySetTagCount(t *testing.T) {
	set := policy.NewPolicySet("release-rules", 1)
	set.Rules = []policy.Rule{{ID: "r1", Metric: "count", RequiredTags: []string{"prod", "signed"}}}
	if got := set.TagCount(); got != 2 {
		t.Fatalf("expected 2 unique tags, got %d", got)
	}
}

func TestEvaluatorRejectsMissingRequiredTag(t *testing.T) {
	set := policy.NewPolicySet("release-rules", 1)
	set.Rules = []policy.Rule{{ID: "r1", Metric: "count", MinEpsilon: 0.1, MaxEpsilon: 10, MaxDelta: 0.01, MinRows: 1, MaxRows: 1000, RequiredTags: []string{"prod"}, Effect: policy.EffectAllow}}
	ev := NewEvaluator(set)
	got, err := ev.Evaluate(Request{Metric: "count", Rows: 10, Epsilon: 1, Delta: 0.001})
	if err != nil {
		t.Fatal(err)
	}
	if got.Allowed {
		t.Fatal("expected missing required tag to be denied")
	}
}

func TestEvaluatorTrimsMetricWhitespace(t *testing.T) {
	set := policy.NewPolicySet("release-rules", 1)
	set.Rules = []policy.Rule{{ID: "r1", Metric: "count", MinEpsilon: 0.1, MaxEpsilon: 10, MaxDelta: 0.01, MinRows: 1, MaxRows: 1000, RequiredTags: []string{"prod"}, Effect: policy.EffectAllow}}
	ev := NewEvaluator(set)
	got, err := ev.Evaluate(Request{Metric: " count ", Rows: 10, Epsilon: 1, Delta: 0.001, Tags: []string{"prod"}})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Allowed {
		t.Fatal("expected metric whitespace to be normalized")
	}
}
