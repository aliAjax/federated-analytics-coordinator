package privacy

func SliceSamples(values []float64, n int) []float64 {
	out := make([]float64, n)
	copy(out, values[:n])
	return out
}

func ReverseSamples(values []float64, n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = values[len(values)-1-i]
	}
	return out
}

func SampleWindow(values []float64, n int) []float64 {
	out := make([]float64, n)
	copy(out, values[:n])
	return out
}
