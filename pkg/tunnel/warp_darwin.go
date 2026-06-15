//go:build darwin

package tunnel

import (
	"fmt"
	"os/exec"
	"strings"
)

// WarpClient implements the Provider interface for Cloudflare WARP.
type WarpClient struct{}

// NewWarpClient initializes the controller.
func NewWarpClient() *WarpClient {
	return &WarpClient{}
}

// Connect engages the WARP tunnel.
func (w *WarpClient) Connect() error {
	exec.Command("warp-cli", "add-excluded-route", "127.0.0.0/8").Run()
	cmd := exec.Command("warp-cli", "connect")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect WARP: %v (Output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// Disconnect disengages the WARP tunnel.
func (w *WarpClient) Disconnect() error {
	cmd := exec.Command("warp-cli", "disconnect")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disconnect WARP: %v (Output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// Status checks if the tunnel is currently active.
func (w *WarpClient) Status() (string, error) {
	cmd := exec.Command("warp-cli", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get WARP status: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
