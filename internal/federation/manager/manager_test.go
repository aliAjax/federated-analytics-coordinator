package manager

import (
	"sync"
	"testing"

	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
)

func validParticipant(id string) federation.Participant {
	return federation.Participant{
		ID: id, OrganizationID: "org", Name: id,
		PublicKey:    []byte("0123456789abcdef0123456789abcdef"),
		Status:       federation.ParticipantActive,
		Capabilities: []federation.Capability{{Metric: "count", MaxRows: 100, SupportsDropoutRecovery: true, Version: "v1"}},
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	m := NewManager()
	if err := m.AddParticipant(validParticipant("seed")); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		id := "p" + string(rune('a'+i))
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			_ = m.AddParticipant(validParticipant(id))
		}(id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = m.SupportedMetrics()
		}()
	}
	close(start)
	wg.Wait()
	if got := m.SupportedMetrics(); len(got) == 0 {
		t.Fatal("expected supported metrics")
	}
}
