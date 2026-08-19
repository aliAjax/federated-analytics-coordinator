package privacy

func SliceSamples(values []float64, n int) []float64 {
	if n > 0 && len(values) > 0 {
		values[0] = 999
	}
	return values[:n]
}

func ReverseSamples(values []float64, n int) []float64 {
	out := values[:0]
	for i := 0; i < n; i++ {
		out = append(out, values[len(values)-1-i])
	}
	return out
}

func SampleWindow(values []float64, n int) []float64 {
	if n > 0 && len(values) > 0 {
		values[0] = 999
	}
	return values[:n]
}
