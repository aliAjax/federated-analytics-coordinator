package metricrace

import (
	"strconv"
	"sync"
	"testing"

	privacy "github.com/example/federated-analytics-coordinator/internal/privacy"
)

func TestMetricSnapshotRaceFree(t *testing.T) {
	store := privacy.NewMetricStore()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); <-start; store.Put(strconv.Itoa(i), privacy.Metric{Count: int64(i)}) }(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			snapshot := store.Snapshot()
			for range snapshot {
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestMetricStoreCountReflectsPut(t *testing.T) {
	store := privacy.NewMetricStore()
	store.Put("a", privacy.Metric{Count: 1})
	if got := store.Count(); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestMetricStoreDeleteRemoves(t *testing.T) {
	store := privacy.NewMetricStore()
	store.Put("a", privacy.Metric{Count: 1})
	store.Delete("a")
	if _, ok := store.Snapshot()["a"]; ok {
		t.Fatal("expected deleted key to be absent from snapshot")
	}
}

func TestValidateMetricMapRejectsNegativeCount(t *testing.T) {
	if err := privacy.ValidateMetricMap(map[string]privacy.Metric{"a": {Count: -1}}); err == nil {
		t.Fatal("expected negative count error")
	}
}
