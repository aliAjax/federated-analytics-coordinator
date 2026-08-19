package domain

import (
	"testing"
	"time"
)

func TestTaskStateMachine(t *testing.T) {
	task := Task{State: Draft}
	if err := task.Transition(Completed, time.Now()); err == nil {
		t.Fatal("draft must not jump to completed")
	}
	if err := task.Transition(Collecting, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestBudgetCannotOverspend(t *testing.T) {
	b := Budget{Epsilon: 1}
	if err := b.Reserve(.8); err != nil {
		t.Fatal(err)
	}
	if err := b.Reserve(.3); err == nil {
		t.Fatal("expected budget exhaustion")
	}
}
