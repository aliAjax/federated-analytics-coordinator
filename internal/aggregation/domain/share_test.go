package domain

import (
	"testing"
	"time"
)

func TestSumRejectsDuplicateParticipants(t *testing.T) {
	_, err := Sum("task", []Share{{TaskID: "task", ParticipantID: "p1", Value: 1}, {TaskID: "task", ParticipantID: "p1", Value: 2}}, time.Now())
	if err == nil {
		t.Fatal("expected duplicate participant rejection")
	}
}
