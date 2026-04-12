package daemon

import (
	"context"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/components/containers"
	zmqclient "github.com/amirhnajafiz/bedrock-api/internal/components/zmq_client"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"go.uber.org/zap"
)

// Daemon represents the main loop of the Docker Daemon that manages system sessions.
type Daemon struct {
	// public shared modules
	ContainerManager containers.ContainerManager
	Logr             *zap.Logger
	APITimeout       time.Duration
	PullInterval     time.Duration

	// private modules
	name        string
	datadir     string
	tracerImage string
	zclient     *zmqclient.ZMQClient
}

// Build initializes the daemon and returns it.
func (d Daemon) Build(name, datadir, tracerImage, apiAddress string) *Daemon {
	d.name = name
	d.datadir = datadir
	d.tracerImage = tracerImage
	d.zclient = zmqclient.NewZMQClient(apiAddress)

	return &d
}

// Serve starts the daemon and polls for sessions from API.
func (d Daemon) Serve(ctx context.Context) error {
	d.Logr.Info("starting daemon", zap.String("name", d.name))

	for {
		// interval between each API call with context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d.PullInterval):
		}

		d.Logr.Info("syncing with API")

		// prepare the pull request events
		events, err := d.prepareEvents()
		if err != nil {
			d.Logr.Warn("failed to prepare pull request events", zap.Error(err))
			continue
		}

		d.Logr.Debug("prepared events", zap.Int("count", len(events)))

		// create a packet with the events to send to the API
		packet := models.NewPacket().WithSender(d.name).WithEvents(events...)

		// send the packet to ZMQ server
		resp, err := d.zclient.SendWithTimeout(packet.ToBytes(), int(d.APITimeout.Seconds()))
		if err != nil {
			d.Logr.Warn("failed to call API", zap.Error(err))
			continue
		}

		// get the response from ZMQ server
		response, err := models.PacketFromBytes(resp)
		if err != nil {
			d.Logr.Warn("failed to parse packet", zap.Error(err))
			continue
		}

		d.Logr.Debug("received events", zap.Int("count", len(response.Events)))

		// sync the local container state with the API state
		ers := d.syncEvents(response.Events)
		for _, er := range ers {
			d.Logr.Warn("failed to sync with API", zap.Error(er))
		}

		d.Logr.Info("synced with API")
	}
}
