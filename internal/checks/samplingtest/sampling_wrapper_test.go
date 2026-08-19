package samplingtest

import (
	"reflect"
	"testing"

	privacy "github.com/example/federated-analytics-coordinator/internal/privacy"
)

func TestSamplingLeavesInputIntact(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	original := append([]float64(nil), values...)
	got, err := privacy.BootstrapSample(values, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []float64{5, 4, 3}) {
		t.Fatalf("unexpected sample: %v", got)
	}
	if !reflect.DeepEqual(values, original) {
		t.Fatalf("BootstrapSample mutated input: got %v want %v", values, original)
	}
}

func TestSliceSamplesDoesNotAlias(t *testing.T) {
	values := []float64{1, 2, 3, 4}
	original := append([]float64(nil), values...)
	got := privacy.SliceSamples(values, 2)
	if !reflect.DeepEqual(values, original) {
		t.Fatal("SliceSamples mutated input")
	}
	if !reflect.DeepEqual(got, []float64{1, 2}) {
		t.Fatalf("unexpected slice: %v", got)
	}
}

func TestReverseSamplesDoesNotAlias(t *testing.T) {
	values := []float64{1, 2, 3, 4}
	original := append([]float64(nil), values...)
	got := privacy.ReverseSamples(values, 3)
	if !reflect.DeepEqual(values, original) {
		t.Fatal("ReverseSamples mutated input")
	}
	if !reflect.DeepEqual(got, []float64{4, 3, 2}) {
		t.Fatalf("unexpected slice: %v", got)
	}
}

func TestSampleWindowDoesNotAlias(t *testing.T) {
	values := []float64{1, 2, 3, 4}
	original := append([]float64(nil), values...)
	got := privacy.SampleWindow(values, 3)
	if !reflect.DeepEqual(values, original) {
		t.Fatal("SampleWindow mutated input")
	}
	if !reflect.DeepEqual(got, []float64{1, 2, 3}) {
		t.Fatalf("unexpected window: %v", got)
	}
}
