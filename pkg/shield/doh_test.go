package shield

import (
	"testing"

	"typo/pkg/logger"

	"github.com/miekg/dns"
)

func TestForwardToDoH(t *testing.T) {
	// 1. Setup our custom logger in DEBUG mode so we can trace the execution
	log := logger.New(logger.LevelDebug)

	// 2. Craft a raw DNS request for google.com (Type A record)
	msg := new(dns.Msg)
	// dns.Fqdn ensures the domain ends with a dot, which is required by the DNS spec
	msg.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)

	// Pack the struct into raw bytes
	rawDNS, err := msg.Pack()
	if err != nil {
		t.Fatalf("Failed to pack DNS message: %v", err)
	}

	// 3. Execution: Send it through our new DoH function
	respBytes, err := ForwardToDoH(rawDNS, DoHCloudflare, log)
	if err != nil {
		t.Fatalf("ForwardToDoH failed: %v", err)
	}

	// 4. Assertion & Unpacking: Turn the raw bytes back into a readable struct
	respMsg := new(dns.Msg)
	err = respMsg.Unpack(respBytes)
	if err != nil {
		t.Fatalf("Failed to unpack DoH response: %v", err)
	}

	// We expect Cloudflare to return at least one IP address for google.com
	if len(respMsg.Answer) == 0 {
		t.Errorf("Expected answers in DoH response, got 0")
	} else {
		log.Info("White-Box Test Passed! Resolved google.com to: %s", respMsg.Answer[0].String())
	}
}
