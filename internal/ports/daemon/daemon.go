package daemon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/components/containers"
	zmqclient "github.com/amirhnajafiz/bedrock-api/internal/components/zmq_client"
	"github.com/amirhnajafiz/bedrock-api/pkg/enums"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"go.uber.org/zap"
)

// Daemon represents the main loop of the Docker Daemon that manages system sessions.
type Daemon struct {
	// public shared modules
	ContainerManager containers.ContainerManager
	Logr             *zap.Logger
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
	for {
		// check if the context is done before each iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// interval between each API call
		time.Sleep(d.PullInterval)

		// get the list of containers
		cts, err := d.ContainerManager.List(context.Background())
		if err != nil {
			d.Logr.Warn("failed to monitor containers", zap.Error(err))
			continue
		}

		// set sessions with containers data
		sessions := make([]models.Session, 0)
		for _, c := range cts {
			status := enums.SessionStatusRunning
			if c.Exited {
				if c.ExitCode == 0 {
					status = enums.SessionStatusFinished
				} else {
					status = enums.SessionStatusFailed
				}
			}

			sessions = append(sessions, models.Session{
				Id:     c.ID,
				Status: status,
			})
		}

		// build a packet with the container sessions
		packet := models.NewPacket().WithSender(d.name).WithSessions(sessions...)

		// send the packet to ZMQ server
		resp, err := d.zclient.SendWithTimeout(packet.ToBytes(), int(d.PullInterval.Seconds()))
		if err != nil {
			d.Logr.Warn("failed to call API", zap.Error(err))
			continue
		}

		// get the response from ZMQ server
		respPacket, err := models.PacketFromBytes(resp)
		if err != nil {
			d.Logr.Warn("failed to parse packet", zap.Error(err))
			continue
		}

		// make changes to reach to API state
		for _, session := range respPacket.Sessions {
			switch session.Status {
			case enums.SessionStatusStopped:
			case enums.SessionStatusFailed:
			case enums.SessionStatusFinished:
				// must stop the target and tracer containers for stopped, failed, and finished sessions
				if err := d.stopContainersForSession(session); err != nil {
					d.Logr.Warn("failed to stop container", zap.String("id", session.Id), zap.Error(err))
				}
			case enums.SessionStatusPending:
				// start the target and tracer containers for pending sessions
				if err := d.startContainersForSession(session); err != nil {
					d.Logr.Warn("failed to start container", zap.String("id", session.Id), zap.Error(err))
				}
			}
		}
	}
}

func (d Daemon) startContainersForSession(session models.Session) error {
	target := fmt.Sprintf("bedrock-target-%s", session.Id)
	tracer := fmt.Sprintf("bedrock-tracer-%s", session.Id)

	// create the output directory for the tracer
	if err := createTracerOutputDir(d.datadir, session.Id); err != nil {
		return fmt.Errorf("failed to create tracer output directory: %w", err)
	}

	// start the tracer container
	if _, err := d.ContainerManager.Start(
		context.Background(),
		&containers.ContainerConfig{
			Name:    tracer,
			Image:   d.tracerImage,
			Cmd:     defaultTracerCommand(target),
			Flags:   defaultContainerFlags(),
			Volumes: defaultTracerVolumes(d.datadir, session.Id),
		},
	); err != nil {
		return err
	}

	// start the target container
	if _, err := d.ContainerManager.Start(
		context.Background(),
		&containers.ContainerConfig{
			Name:  target,
			Image: session.Spec.Image,
			Cmd:   strings.Split(session.Spec.Command, " "),
		},
	); err != nil {
		return err
	}

	return nil
}

func (d Daemon) stopContainersForSession(session models.Session) error {
	// create a new background context for stopping containers
	ctx := context.Background()

	target := fmt.Sprintf("bedrock-target-%s", session.Id)
	tracer := fmt.Sprintf("bedrock-tracer-%s", session.Id)

	// check if the target container is running before trying to stop it
	targetInfo, err := d.ContainerManager.Get(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to get container info for %s: %w", target, err)
	}

	if !targetInfo.Exited {
		// stop the target container
		if err := d.ContainerManager.Stop(ctx, target); err != nil {
			return fmt.Errorf("failed to stop container %s: %w", target, err)
		}

		// stop the tracer container
		if err := d.ContainerManager.Stop(ctx, tracer); err != nil {
			return fmt.Errorf("failed to stop container %s: %w", tracer, err)
		}
	}

	// remove both containers after stopping
	if err := d.ContainerManager.Remove(ctx, target); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", target, err)
	}
	if err := d.ContainerManager.Remove(ctx, tracer); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", tracer, err)
	}

	return nil
}
