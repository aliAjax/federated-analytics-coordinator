package application

import (
	"fmt"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func (s *Service) CanComplete(t protocol.Task) bool {
	return t.State == protocol.Revealing && t.Budget.Spent <= t.Budget.Epsilon
}

func (s *Service) CompletionReason(t protocol.Task) string {
	if t.Budget.Spent > t.Budget.Epsilon {
		return "budget overspent"
	}
	if t.State != protocol.Revealing {
		return "task not revealing"
	}
	return "ok"
}

func (s *Service) CompletionAllowed(b protocol.Budget) error {
	if b.Epsilon <= 0 {
		return fmt.Errorf("task budget must be positive")
	}
	if b.Spent < 0 {
		return fmt.Errorf("spent budget cannot be negative")
	}
	if b.Delta < 0 {
		return fmt.Errorf("delta cannot be negative")
	}
	return nil
}
