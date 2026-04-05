package scheduler_test

import (
	"testing"

	"github.com/amirhnajafiz/bedrock-api/internal/scheduler"
)

// TestRoundRobinScheduler tests the functionality of the RoundRobin scheduler to ensure it correctly
// appends, picks, and drops items in a round-robin manner.
func TestRoundRobinScheduler(t *testing.T) {
	// create a new round-robin scheduler
	sc := scheduler.NewRoundRobin()

	// test appending items to the scheduler
	sc.Append("item1")
	sc.Append("item2")
	sc.Append("item3")

	// test the order of items in the scheduler
	if item, err := sc.Pick(); err != nil || item != "item1" {
		t.Errorf("Expected item1, got %s", item)
	}
	if item, err := sc.Pick(); err != nil || item != "item2" {
		t.Errorf("Expected item2, got %s", item)
	}
	if item, err := sc.Pick(); err != nil || item != "item3" {
		t.Errorf("Expected item3, got %s", item)
	}

	// test dropping an item from the scheduler
	sc.Drop("item2")

	// test the order of items after dropping one
	if item, err := sc.Pick(); err != nil || item != "item1" {
		t.Errorf("Expected item1, got %s", item)
	}
	if item, err := sc.Pick(); err != nil || item != "item3" {
		t.Errorf("Expected item3, got %s", item)
	}
	if item, err := sc.Pick(); err != nil || item != "item1" {
		t.Errorf("Expected item1, got %s", item)
	}

	// test dropping all items from the scheduler
	sc.Drop("item1")
	sc.Drop("item3")

	// test picking from an empty scheduler
	if _, err := sc.Pick(); err == nil {
		t.Errorf("Expected error when picking from an empty scheduler")
	}

	// test appending duplicate items to the scheduler
	sc.Append("item1")
	sc.Append("item1")
	sc.Append("item2")

	// test the order of items with duplicates
	if item, err := sc.Pick(); err != nil || item != "item1" {
		t.Errorf("Expected item1, got %s", item)
	}
	if item, err := sc.Pick(); err != nil || item != "item2" {
		t.Errorf("Expected item2, got %s", item)
	}
}
