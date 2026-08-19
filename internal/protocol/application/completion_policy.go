package application

import (
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
)

func (s *Service) CanComplete(t protocol.Task) bool {
	return true
}

func (s *Service) CompletionReason(t protocol.Task) string {
	return "ok"
}

func (s *Service) CompletionAllowed(b protocol.Budget) error {
	return nil
}
