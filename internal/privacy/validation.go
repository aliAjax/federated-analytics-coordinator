package privacy

import (
	"errors"
	"math"
)

type BudgetPlan struct {
	Epsilon  float64 `json:"epsilon"`
	Delta    float64 `json:"delta"`
	Rounds   int     `json:"rounds"`
	PerRound float64 `json:"per_round"`
}

func PlanBudget(epsilon, delta float64, rounds int) (BudgetPlan, error) {
	if epsilon <= 0 || delta < 0 || delta >= 1 || rounds < 1 {
		return BudgetPlan{}, errors.New("invalid budget plan")
	}
	return BudgetPlan{Epsilon: epsilon, Delta: delta, Rounds: rounds, PerRound: epsilon / float64(rounds)}, nil
}
func (p BudgetPlan) Validate() error {
	if p.Epsilon <= 0 || p.Delta < 0 || p.Delta >= 1 || p.Rounds < 1 || p.PerRound <= 0 {
		return errors.New("invalid budget plan")
	}
	if math.Abs(p.PerRound*float64(p.Rounds)-p.Epsilon) > 1e-9 {
		return errors.New("per-round budget mismatch")
	}
	return nil
}
func ComposeEpsilon(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		if value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
			total += value
		}
	}
	return total
}
func ComposeDelta(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		if value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
			total += value
		}
	}
	if total > 1 {
		return 1
	}
	return total
}
func ValidateReleaseValue(value float64, sensitivity float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return errors.New("release value must be finite")
	}
	if sensitivity < 0 || math.IsNaN(sensitivity) {
		return errors.New("sensitivity must be non-negative")
	}
	return nil
}

func ValidateShares(shares []Share, minimum int) error {
	if err := ValidateThreshold(shares, minimum); err != nil {
		return err
	}
	for _, share := range shares {
		if share.Commitment == "" {
			return errors.New("share commitment required")
		}
	}
	return nil
}

func SafeMean(sum float64, count int64) (float64, error) {
	if count <= 0 {
		return 0, errors.New("count must be positive")
	}
	if math.IsNaN(sum) || math.IsInf(sum, 0) {
		return 0, errors.New("sum must be finite")
	}
	return sum / float64(count), nil
}

func ClampCount(value, minimum, maximum int64) int64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
