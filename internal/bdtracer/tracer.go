package bdtracer

import (
	"fmt"
	"os"
)

const (
	BaseDir      = "/tmp/bedrock-outputs"
	BdtraceImage = "ghcr.io/amirhnajafiz/bedrock-tracer:v0.0.6-beta"
)

func CreateTracerOutputDir(sessionId string) error {
	outputDir := fmt.Sprintf("%s/%s", BaseDir, sessionId)
	return os.MkdirAll(outputDir, 0755)
}

func DefaultContainerFlags() map[string]any {
	return map[string]any{
		"pid":        "host",
		"privileged": true,
	}
}

func DefaultTracerVolumes(sessionId string) map[string]string {
	return map[string]string{
		"/sys":                    "/sys:rw",
		"/lib/modules":            "/lib/modules:ro",
		"/var/run/docker.sock":    "/var/run/docker.sock",
		BaseDir + "/" + sessionId: "/logs",
	}
}

func DefaultTracerCommand(targetContainerName string) []string {
	return []string{
		"bdtrace",
		"--container",
		targetContainerName,
		"-o",
		"/logs",
	}
}
