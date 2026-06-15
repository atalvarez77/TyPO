package warden

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const AppleCaptiveURL = "http://captive.apple.com/hotspot-detect.html"

// IsTrapped returns true if the network is intercepting HTTP traffic
func IsTrapped() bool {
	// Create a custom client with a strict timeout
	// Captive portals often intentionally delay responses
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(AppleCaptiveURL)
	if err != nil {
		// A total connection failure usually means a portal is dropping packets
		return true
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return true
	}

	// Apple's server always returns a simple HTML string containing "Success"
	// If the portal intercepted us, we will get their login HTML instead
	return !strings.Contains(string(bodyBytes), "Success")
}
