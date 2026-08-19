package engine

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	aggregate "github.com/example/federated-analytics-coordinator/internal/aggregation/domain"
)

// Engine turns validated shares into protocol-level aggregate views.
type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

type Quantile struct {
	P     float64 `json:"p"`
	Value float64 `json:"value"`
}

type Histogram struct {
	Buckets []float64 `json:"buckets"`
	Counts  []int64   `json:"counts"`
}

// WeightedQuantile computes interpolated quantiles from already sorted values.
func WeightedQuantile(values []float64, weights []float64, ps []float64) ([]Quantile, error) {
	if len(values) == 0 {
		return nil, errors.New("values are required")
	}
	if len(values) != len(weights) {
		return nil, errors.New("values and weights must have equal length")
	}
	total := 0.0
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("value at %d is not finite", i)
		}
		if math.IsNaN(weights[i]) || math.IsInf(weights[i], 0) || weights[i] < 0 {
			return nil, fmt.Errorf("weight at %d is invalid", i)
		}
		total += weights[i]
	}
	if total <= 0 {
		return nil, errors.New("weights must have positive sum")
	}
	ordered := make([]struct {
		value  float64
		weight float64
	}, len(values))
	for i := range values {
		ordered[i].value = values[i]
		ordered[i].weight = weights[i]
	}
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].value < ordered[j].value })
	out := make([]Quantile, 0, len(ps))
	for _, p := range ps {
		if math.IsNaN(p) || p < 0 || p > 1 {
			return nil, fmt.Errorf("invalid quantile %v", p)
		}
		target := p * total
		acc := 0.0
		for _, item := range ordered {
			acc += item.weight
			if acc >= target {
				out = append(out, Quantile{P: p, Value: item.value})
				break
			}
		}
		if len(out) == len(ps) {
			continue
		}
		out = append(out, Quantile{P: p, Value: ordered[len(ordered)-1].value})
	}
	return out, nil
}

// Histogramize puts each value in the first bucket whose right edge is >= value.
func Histogramize(values, buckets []float64) (Histogram, error) {
	if len(buckets) == 0 {
		return Histogram{}, errors.New("buckets are required")
	}
	edges := append([]float64(nil), buckets...)
	sort.Float64s(edges)
	for i, v := range edges {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return Histogram{}, fmt.Errorf("bucket at %d is not finite", i)
		}
	}
	for i := 1; i < len(edges); i++ {
		if edges[i] <= edges[i-1] {
			return Histogram{}, errors.New("buckets must be strictly increasing")
		}
	}
	counts := make([]int64, len(edges)+1)
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return Histogram{}, fmt.Errorf("value is not finite")
		}
		idx := sort.SearchFloat64s(edges, v)
		counts[idx]++
	}
	return Histogram{Buckets: edges, Counts: counts}, nil
}

// MergeShares validates and sums shares, preserving deterministic participant order.
func (e *Engine) MergeShares(taskID string, shares []aggregate.Share) (aggregate.Aggregate, error) {
	if taskID == "" {
		return aggregate.Aggregate{}, errors.New("task id is required")
	}
	shares = aggregate.CompactShares(shares)
	return aggregate.Sum(taskID, shares, time.Now().UTC())
}

func (e *Engine) ParticipantCount(shares []aggregate.Share) int {
	count := 0
	for _, s := range shares {
		if s.ParticipantID != "" {
			count++
		}
	}
	return count
}
