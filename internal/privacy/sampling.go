package privacy

import (
	"errors"
	"math"
	"sort"
)

type Quantile struct {
	P     float64 `json:"p"`
	Value float64 `json:"value"`
}

func Quantiles(values []float64, ps []float64) ([]Quantile, error) {
	if len(values) == 0 {
		return nil, errors.New("values are required")
	}
	clean := make([]float64, 0, len(values))
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("values must be finite")
		}
		clean = append(clean, v)
	}
	sort.Float64s(clean)
	out := make([]Quantile, 0, len(ps))
	for _, p := range ps {
		if p < 0 || p > 1 {
			return nil, errors.New("quantile must be between zero and one")
		}
		index := int(math.Round(p * float64(len(clean)-1)))
		out = append(out, Quantile{P: p, Value: clean[index]})
	}
	return out, nil
}

type Histogram struct {
	Buckets []float64 `json:"buckets"`
	Counts  []int64   `json:"counts"`
}

func Histogramize(values, buckets []float64) (Histogram, error) {
	if len(buckets) == 0 {
		return Histogram{}, errors.New("buckets are required")
	}
	sorted := append([]float64(nil), buckets...)
	sort.Float64s(sorted)
	counts := make([]int64, len(sorted)+1)
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return Histogram{}, errors.New("value must be finite")
		}
		index := sort.SearchFloat64s(sorted, v)
		counts[index]++
	}
	return Histogram{Buckets: sorted, Counts: counts}, nil
}
func MergeHistograms(items []Histogram) (Histogram, error) {
	if len(items) == 0 {
		return Histogram{}, errors.New("histograms are required")
	}
	base := append([]float64(nil), items[0].Buckets...)
	counts := make([]int64, len(items[0].Counts))
	for _, item := range items {
		if len(item.Buckets) != len(base) || len(item.Counts) != len(counts) {
			return Histogram{}, errors.New("histogram shape mismatch")
		}
		for i, v := range item.Buckets {
			if v != base[i] {
				return Histogram{}, errors.New("histogram bucket mismatch")
			}
		}
		for i, v := range item.Counts {
			if v < 0 {
				return Histogram{}, errors.New("negative count")
			}
			counts[i] += v
		}
	}
	return Histogram{Buckets: base, Counts: counts}, nil
}

func AddHistogramNoise(hist Histogram, mechanism Mechanism, epsilon, delta, sensitivity float64) (Histogram, error) {
	if epsilon <= 0 || sensitivity < 0 {
		return Histogram{}, errors.New("invalid noise budget")
	}
	out := Histogram{Buckets: append([]float64(nil), hist.Buckets...), Counts: append([]int64(nil), hist.Counts...)}
	for i, count := range hist.Counts {
		var noise float64
		var err error
		switch mechanism {
		case NoNoise:
		case Laplace:
			noise, err = LaplaceNoise(sensitivity, epsilon)
		case Gaussian:
			noise, err = GaussianNoise(sensitivity, epsilon, delta)
		default:
			return Histogram{}, errors.New("unsupported mechanism")
		}
		if err != nil {
			return Histogram{}, err
		}
		value := float64(count) + noise
		if value < 0 {
			value = 0
		}
		out.Counts[i] = int64(math.Round(value))
	}
	return out, nil
}

func BootstrapSample(values []float64, n int) ([]float64, error) {
	if n < 0 || n > len(values) {
		return nil, errors.New("invalid sample size")
	}
	out := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, values[len(values)-1-i])
	}
	return out, nil
}
