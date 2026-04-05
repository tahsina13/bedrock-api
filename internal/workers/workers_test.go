package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/scheduler"
	"github.com/amirhnajafiz/bedrock-api/internal/workers"
)

// TestWorkerDockerDHealthCheck tests the WorkerDockerDHealthCheck function to ensure it correctly updates
// the scheduler with healthy Docker daemons and removes stale entries after the specified interval.
func TestWorkerDockerDHealthCheck(t *testing.T) {
	// create a context
	ctx, cancel := context.WithCancel(context.Background())

	// create a channel to simulate Docker daemon health updates
	input := make(chan string)

	// get a reference to the scheduler instance
	sc := scheduler.NewRoundRobin()

	// start the WorkerDockerDHealthCheck in a separate goroutine
	go workers.WorkerDockerDHealthCheck(ctx, input, 3*time.Second)

	// simulate sending health updates for Docker daemons
	input <- "dockerd1"
	input <- "dockerd2"

	time.Sleep(1 * time.Second)

	// check if the scheduler has the expected Docker daemons
	if !sc.Exists("dockerd1") {
		t.Errorf("Expected dockerd1 to be in the scheduler")
	}
	if !sc.Exists("dockerd2") {
		t.Errorf("Expected dockerd2 to be in the scheduler")
	}

	// wait for a short period to allow the worker to process the updates
	time.Sleep(5 * time.Second)

	// check if the scheduler has removed the stale Docker daemons
	if sc.Exists("dockerd1") {
		t.Errorf("Expected dockerd1 to be removed from the scheduler")
	}
	if sc.Exists("dockerd2") {
		t.Errorf("Expected dockerd2 to be removed from the scheduler")
	}

	// cancel the context to stop the worker
	cancel()
}
