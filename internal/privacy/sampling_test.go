package privacy

import (
	"reflect"
	"testing"
)

func TestBootstrapSampleDoesNotMutateInput(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	original := append([]float64(nil), values...)
	got, err := BootstrapSample(values, 3)
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
