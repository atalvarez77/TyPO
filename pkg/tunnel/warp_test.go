package tunnel

import (
	"strings"
	"testing"
	"time"
)

func TestWarpClient(t *testing.T) {
	client := NewWarpClient()

	// Connect to the WARP edge network
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect WARP: %v", err)
	}

	// The cryptographic handshake takes a moment to establish
	time.Sleep(2 * time.Second)

	// Verify the daemon reports an active connection
	status, err := client.Status()
	if err != nil {
		t.Fatalf("Failed to get WARP status: %v", err)
	}

	if !strings.Contains(status, "Connected") {
		t.Errorf("Expected status to contain 'Connected', got: %s", status)
	}

	// Disconnect to restore the original local network state
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect WARP: %v", err)
	}
}
