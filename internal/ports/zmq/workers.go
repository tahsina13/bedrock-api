package zmq

import (
	"context"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/scheduler"
)

// workerCheckDockerDHealthStatus continuously checks the health status of Docker daemons by listening to a healthChannel for updates and using
// a ticker to periodically remove stale entries from the health map.
// It interacts with the scheduler to append or drop Docker daemons based on their health status.
func workerCheckDockerDHealthStatus(ctx context.Context, healthChannel chan string, interval time.Duration, scheduler scheduler.Scheduler) {
	// healthMap keeps track of the last time a health update was received for each Docker daemon
	healthMap := make(map[string]time.Time)

	// ticker is used to periodically check for stale entries in the healthMap
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case dockerd := <-healthChannel:
			// update the healthMap with the current time for the received Docker daemon
			healthMap[dockerd] = time.Now()
			scheduler.Append(dockerd)
		case <-ticker.C:
			timeSnapshot := time.Now()

			// loop through the healthMap and remove any entries that haven't been updated within the interval
			for dockerd, lastUpdated := range healthMap {
				if timeSnapshot.Sub(lastUpdated) > interval {
					delete(healthMap, dockerd)
					scheduler.Drop(dockerd)
				}
			}
		}
	}
}
