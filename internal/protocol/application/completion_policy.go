package application

import (
	"fmt"

	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func (s *Service) CanComplete(t protocol.Task) bool {
	return t.Budget.Spent <= t.Budget.Epsilon
}

func (s *Service) CompletionReason(t protocol.Task) string {
	if t.Budget.Spent > t.Budget.Epsilon {
		return "budget overspent"
	}
	return "ok"
}

func (s *Service) CompletionAllowed(b protocol.Budget) error {
	if b.Delta < 0 {
		return fmt.Errorf("budget delta must not be negative")
	}
	return nil
}
