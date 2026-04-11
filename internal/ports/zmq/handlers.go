package zmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/amirhnajafiz/bedrock-api/pkg/enums"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"github.com/zeromq/goczmq"
	"go.uber.org/zap"
)

// socket receiver reads input messages from router and sends them over handler channel.
func (z ZMQServer) socketReceiver(ctx context.Context, router *goczmq.Sock, channel chan [][]byte) error {
	// set receive timeout to 2 seconds to allow graceful shutdown
	router.SetRcvtimeo(2000)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive message from router
		request, err := router.RecvMessage()
		if err != nil {
			if !errors.Is(err, goczmq.ErrTimeout) && !errors.Is(err, goczmq.ErrRecvFrame) {
				z.Logr.Warn("failed to receive message", zap.Error(err))
			}

			continue
		}

		z.Logr.Debug("new message", zap.String("ip", string(request[0])))

		channel <- request
	}
}

// socket sender reads input from handler channel and sends them to router.
func (z ZMQServer) socketSender(ctx context.Context, router *goczmq.Sock, channel chan [][]byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-channel:
			if err := router.SendMessage(event); err != nil {
				z.Logr.Warn("failed to send message", zap.Error(err))
				return fmt.Errorf("sender router failed: %v", err)
			}
		}

	}
}

// socket handler is the main loop of ZMQ server.
func (z ZMQServer) socketHandler(ctx context.Context, in chan [][]byte, out chan [][]byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-in:
			out <- z.processEvent(event)
		}
	}
}

// process event is the main logic of ZMQ server. It processes incoming events.
func (z ZMQServer) processEvent(event [][]byte) [][]byte {
	// parse events into packets
	pkt, err := models.PacketFromBytes(event[1])
	if err != nil {
		z.Logr.Warn("failed to parse event", zap.Error(err))
		return event
	}

	// reply empty packets
	if pkt.IsEmpty() {
		return [][]byte{event[0], pkt.ToBytes()}
	}

	// create a response packet
	responsePkt := models.NewPacket()
	responsePkt.WithSender("api")

	// check sender header and registration status, if invalid, reply with empty packet
	dockerd := ""
	if val, ok := pkt.Headers["sender"]; !ok {
		z.Logr.Warn("sender header is missing")
		return [][]byte{event[0], responsePkt.ToBytes()}
	} else {
		dockerd = val
	}

	// update health status of the sender daemon
	z.DockerDHealthChannel <- dockerd
	z.Logr.Debug("new message from daemon", zap.String("dockerd", dockerd))

	// read events from packet and update KV storage
	for _, event := range pkt.Events {
		// api only cares about session running, failed, and finished events, skip other events
		if event.GetEventType() != enums.EventTypeSessionRunning && event.GetEventType() != enums.EventTypeSessionFailed && event.GetEventType() != enums.EventTypeSessionEnd {
			continue
		}

		// get the session id from header
		sid := event.GetSessionId()

		// retrieve the session record from KV storage using session id
		// if session record is not found or dockerd id does not match, skip the event
		record, err := z.sessionStore.GetSessionById(sid)
		if err != nil || record.DockerDId != dockerd {
			z.Logr.Warn(
				"failed to get session",
				zap.Error(err),
				zap.String("session id", sid),
				zap.String("dockerd id", dockerd),
			)
			continue
		}

		// set session status to failed and update the timestamp
		if event.GetEventType() == enums.EventTypeSessionFailed {
			record.Status = z.stateMachine.Transition(record.Status, enums.SessionStatusFailed)
		} else if event.GetEventType() == enums.EventTypeSessionRunning {
			record.Status = z.stateMachine.Transition(record.Status, enums.SessionStatusRunning)
		} else if event.GetEventType() == enums.EventTypeSessionEnd {
			record.Status = z.stateMachine.Transition(record.Status, enums.SessionStatusFinished)
		}

		// update the session in KV storage
		if err := z.sessionStore.SaveSession(record); err != nil {
			z.Logr.Warn(
				"failed to update session",
				zap.Error(err),
				zap.String("session id", record.Id),
				zap.String("dockerd id", dockerd),
			)
			continue
		}
	}

	// retrieve all sessions of the sender daemon from KV storage
	sessions, err := z.sessionStore.ListSessionsByDockerDId(dockerd)
	if err != nil {
		z.Logr.Warn("failed to list sessions", zap.Error(err))

		return [][]byte{event[0], responsePkt.ToBytes()}
	}

	// only include running, stopped, or finished sessions
	for _, session := range sessions {
		switch session.Status {
		case enums.SessionStatusPending:
			responsePkt.WithEvents(
				models.NewEvent().
					WithSessionId(session.Id).
					WithEventType(enums.EventTypeSessionStart).
					WithPayload(session.Spec),
			)
		case enums.SessionStatusFailed:
		case enums.SessionStatusStopped:
		case enums.SessionStatusFinished:
			responsePkt.WithEvents(
				models.NewEvent().
					WithSessionId(session.Id).
					WithEventType(enums.EventTypeSessionEnd).
					WithPayload(nil),
			)
		}
	}

	// send the response packet back to the sender
	return [][]byte{event[0], responsePkt.ToBytes()}
}
