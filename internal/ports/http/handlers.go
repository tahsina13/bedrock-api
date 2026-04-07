package http

import (
	"net/http"
	"time"

	"github.com/amirhnajafiz/bedrock-api/pkg/enums"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

// health checks the server's health by sending an empty packet to ZMQ server.
func (h HTTPServer) health(c *echo.Context) error {
	// call the ZMQ server to check if it's alive
	_, err := h.zclient.Send(models.NewPacket().ToBytes())
	if err != nil {
		h.Logr.Warn("zmq server connection error", zap.Error(err))
		return c.String(http.StatusInternalServerError, "zmq not healthy")
	}

	return c.String(http.StatusOK, "OK")
}

// createSession creates a new session based on the request payload and returns the session ID.
func (h HTTPServer) createSession(c *echo.Context) error {
	var spec models.Spec
	if err := c.Bind(&spec); err != nil {
		return c.String(http.StatusBadRequest, "invalid request body")
	}

	// assign DockerD to this session
	dockerd, err := h.scheduler.Pick()
	if err != nil {
		return c.String(http.StatusServiceUnavailable, "no docker daemons available")
	}

	// create and save session to KV store
	session := models.Session{
		Id:        uuid.New().String(),
		DockerDId: dockerd,
		CreatedAt: time.Now(),
		Status:    enums.SessionStatusPending,
		Spec:      spec,
	}
	h.sessionStore.SaveSession(session.Id, session.DockerDId, &session)

	return c.JSON(http.StatusCreated, session)
}

// updateSession updates an existing session with the specified ID based on the request payload.
func (h HTTPServer) updateSession(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

// getSessions retrieves a list of all sessions and returns them in the response.
func (h HTTPServer) getSessions(c *echo.Context) error {
	sessions, err := h.sessionStore.ListSessions()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to list sessions")
	}
	return c.JSON(http.StatusFound, sessions)
}

// getSessionLogs retrieves the logs for a specific session based on the session ID provided in the request parameters.
func (h HTTPServer) getSessionLogs(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}

// storeSessionLogs stores the logs for a specific session based on the session ID provided in the request parameters.
func (h HTTPServer) storeSessionLogs(c *echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}
