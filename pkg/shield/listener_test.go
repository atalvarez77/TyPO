package shield

import (
	"testing"

	"typo/pkg/logger"
	"typo/pkg/sinkhole"

	"github.com/miekg/dns"
)

type mockResponseWriter struct {
	dns.ResponseWriter
	writtenMsg *dns.Msg
}

func (m *mockResponseWriter) WriteMsg(msg *dns.Msg) error {
	m.writtenMsg = msg
	return nil
}

func TestLocalInterceptor_ServeDNS(t *testing.T) {
	log := logger.New(logger.LevelDebug)

	interceptor := NewInterceptor(log, nil)
	// Inject a mock sinkhole directly for the test
	blocklist := sinkhole.NewBlocklist()
	blocklist["ads.example.com"] = true
	interceptor.Sinkhole = &blocklist

	tests := []struct {
		name          string
		domain        string
		expectBlocked bool
	}{
		{"Clean Domain", "cloudflare.com.", false},
		{"Blocked Domain", "ads.example.com.", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := new(dns.Msg)
			req.SetQuestion(tc.domain, dns.TypeA)
			mockW := &mockResponseWriter{}

			interceptor.ServeDNS(mockW, req)

			if tc.expectBlocked {
				// HandleFailed writes a ServerFailure code back to the client
				if mockW.writtenMsg == nil || mockW.writtenMsg.Rcode != dns.RcodeServerFailure {
					t.Errorf("Expected ServerFailure for blocked domain, got invalid response")
				}
			} else {
				if mockW.writtenMsg == nil {
					t.Fatalf("Expected a response for clean domain, got nil")
				}
				if len(mockW.writtenMsg.Answer) == 0 {
					t.Errorf("Expected answers for clean domain, got 0")
				}
			}
		})
	}
}
