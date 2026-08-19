package privacy

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
)

type Mechanism string

const (
	Laplace  Mechanism = "laplace"
	Gaussian Mechanism = "gaussian"
	NoNoise  Mechanism = "none"
)

type Release struct {
	ID          string    `json:"id"`
	Mechanism   Mechanism `json:"mechanism"`
	Epsilon     float64   `json:"epsilon"`
	Delta       float64   `json:"delta"`
	Sensitivity float64   `json:"sensitivity"`
	NoiseScale  float64   `json:"noise_scale"`
	Value       float64   `json:"value"`
}
type Accountant struct {
	mu           sync.Mutex
	epsilon      float64
	delta        float64
	spentEpsilon float64
	spentDelta   float64
	releases     []Release
}

func NewAccountant(epsilon, delta float64) *Accountant {
	return &Accountant{epsilon: epsilon, delta: delta, releases: []Release{}}
}
func (a *Accountant) Reserve(epsilon, delta float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if epsilon <= 0 || delta < 0 {
		return errors.New("invalid privacy budget")
	}
	if a.spentEpsilon+epsilon > a.epsilon || a.spentDelta+delta > a.delta {
		return errors.New("privacy budget exhausted")
	}
	a.spentEpsilon += epsilon
	a.spentDelta += delta
	return nil
}
func (a *Accountant) Release(epsilon, delta float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.spentEpsilon = math.Max(0, a.spentEpsilon-epsilon)
	a.spentDelta = math.Max(0, a.spentDelta-delta)
}
func (a *Accountant) Record(r Release) error {
	if r.ID == "" || r.Epsilon <= 0 || r.Sensitivity < 0 {
		return errors.New("invalid release")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releases = append(a.releases, r)
	return nil
}
func (a *Accountant) Summary() map[string]float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return map[string]float64{"epsilon_limit": a.epsilon, "delta_limit": a.delta, "epsilon_spent": a.spentEpsilon, "delta_spent": a.spentDelta, "remaining_epsilon": math.Max(0, a.epsilon-a.spentEpsilon), "remaining_delta": math.Max(0, a.delta-a.spentDelta)}
}
func (a *Accountant) Releases() []Release {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]Release(nil), a.releases...)
}

func Clamp(value, lower, upper float64) float64 {
	if lower > upper {
		lower, upper = upper, lower
	}
	return math.Min(upper, math.Max(lower, value))
}
func LaplaceNoise(sensitivity, epsilon float64) (float64, error) {
	if sensitivity < 0 || epsilon <= 0 {
		return 0, errors.New("invalid laplace parameters")
	}
	u, err := secureUnit()
	if err != nil {
		return 0, err
	}
	scale := sensitivity / epsilon
	if u < .5 {
		return scale * math.Log(2*u), nil
	}
	return -scale * math.Log(2*(1-u)), nil
}
func GaussianNoise(sensitivity, epsilon, delta float64) (float64, error) {
	if sensitivity < 0 || epsilon <= 0 || delta <= 0 || delta >= 1 {
		return 0, errors.New("invalid gaussian parameters")
	}
	u1, e := secureUnit()
	if e != nil {
		return 0, e
	}
	u2, e := secureUnit()
	if e != nil {
		return 0, e
	}
	sigma := sensitivity * math.Sqrt(2*math.Log(1.25/delta)) / epsilon
	return sigma * math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2), nil
}
func secureUnit() (float64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	n := binary.BigEndian.Uint64(b[:]) >> 11
	return (float64(n) + .5) / float64(uint64(1)<<53), nil
}

type Share struct {
	Participant string `json:"participant"`
	Value       int64  `json:"value"`
	Commitment  string `json:"commitment"`
}

func Commit(shares []Share) string {
	ordered := append([]Share(nil), shares...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Participant < ordered[j].Participant })
	h := sha256.New()
	for _, s := range ordered {
		fmt.Fprintf(h, "%s:%d:%s|", s.Participant, s.Value, s.Commitment)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
func ValidateThreshold(shares []Share, minimum int) error {
	if minimum < 1 {
		return errors.New("minimum must be positive")
	}
	seen := map[string]bool{}
	for _, s := range shares {
		if s.Participant == "" || seen[s.Participant] {
			return errors.New("duplicate or empty participant")
		}
		seen[s.Participant] = true
	}
	if len(shares) < minimum {
		return fmt.Errorf("threshold %d not met by %d shares", minimum, len(shares))
	}
	return nil
}
func SumShares(shares []Share) (int64, error) {
	var total int64
	for _, s := range shares {
		if (s.Value > 0 && total > math.MaxInt64-s.Value) || (s.Value < 0 && total < math.MinInt64-s.Value) {
			return 0, errors.New("share sum overflow")
		}
		total += s.Value
	}
	return total, nil
}

func (a *Accountant) Settle(l *Ledger, entry LedgerEntry) (err error) {
	if err = a.Reserve(entry.Epsilon, entry.Delta); err != nil {
		return err
	}
	if err = l.Reserve(entry); err != nil {
		a.refundBudget(entry)
		return err
	}
	if err = validateLedgerDelta(entry.Delta); err != nil {
		a.refundBudget(entry)
		return err
	}
	if err = l.Commit(entry.ID); err != nil {
		a.refundBudget(entry)
		return err
	}
	return nil
}

func (a *Accountant) refundBudget(entry LedgerEntry) {
	a.Release(entry.Epsilon, entry.Delta)
}

func validateLedgerDelta(delta float64) error {
	if delta >= 1 {
		return errors.New("delta must be below one")
	}
	if delta < 0 {
		return errors.New("delta cannot be negative")
	}
	return nil
}
