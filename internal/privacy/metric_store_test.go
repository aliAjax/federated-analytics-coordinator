package privacy

import (
	"strconv"
	"sync"
	"testing"
)

func TestMetricStoreConcurrentSnapshot(t *testing.T) {
	store := NewMetricStore()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			store.Put(strconv.Itoa(i), Metric{Count: int64(i)})
		}(i)
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
