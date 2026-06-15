package shield

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"typo/pkg/logger"
)

// DoHCloudflare is the standard endpoint for Cloudflare's 1.1.1.1 DoH service.
const DoHCloudflare = "https://1.1.1.1/dns-query"

// ForwardToDoH takes a raw DNS UDP payload, wraps it in an HTTPS request,
// sends it to a DoH provider, and returns the raw DNS response payload.
func ForwardToDoH(rawDNS []byte, upstreamURL string, log *logger.Logger) ([]byte, error) {
	log.Debug("Preparing DoH payload (%d bytes) for upstream: %s", len(rawDNS), upstreamURL)

	// 1. Wrap the raw DNS bytes into an HTTP POST request
	req, err := http.NewRequest(http.MethodPost, upstreamURL, bytes.NewReader(rawDNS))
	if err != nil {
		return nil, fmt.Errorf("failed to create DoH request: %w", err)
	}

	// 2. Set the strict headers required by the RFC 8484 DoH specification
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	// 3. Dispatch the encrypted request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream DoH request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Validate the response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned non-200 status: %d", resp.StatusCode)
	}

	// 5. Read the response payload back into raw bytes
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DoH response body: %w", err)
	}

	log.Debug("Successfully received DoH response (%d bytes)", len(respBytes))
	return respBytes, nil
}
