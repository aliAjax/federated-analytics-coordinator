package privacy

import (
	"errors"
	"math"
	"sync"
)

type Metric struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
}
type MetricAccumulator struct {
	mu          sync.Mutex
	count       int64
	sum         float64
	min         float64
	max         float64
	initialized bool
}

func NewMetricAccumulator() *MetricAccumulator { return &MetricAccumulator{} }
func (a *MetricAccumulator) Add(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return errors.New("metric value must be finite")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.initialized {
		a.min = value
		a.max = value
		a.initialized = true
	} else {
		if value < a.min {
			a.min = value
		}
		if value > a.max {
			a.max = value
		}
	}
	a.count++
	a.sum += value
	return nil
}
func (a *MetricAccumulator) Merge(other Metric) error {
	if other.Count < 0 || math.IsNaN(other.Sum) || math.IsInf(other.Sum, 0) {
		return errors.New("invalid metric")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if other.Count == 0 {
		return nil
	}
	if !a.initialized {
		a.min = other.Min
		a.max = other.Max
		a.initialized = true
	} else {
		if other.Min < a.min {
			a.min = other.Min
		}
		if other.Max > a.max {
			a.max = other.Max
		}
	}
	a.count += other.Count
	a.sum += other.Sum
	return nil
}
func (a *MetricAccumulator) Snapshot() Metric {
	a.mu.Lock()
	defer a.mu.Unlock()
	m := Metric{Count: a.count, Sum: a.sum, Min: a.min, Max: a.max}
	if a.count > 0 {
		m.Mean = a.sum / float64(a.count)
	}
	return m
}
func (a *MetricAccumulator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count = 0
	a.sum = 0
	a.min = 0
	a.max = 0
	a.initialized = false
}
