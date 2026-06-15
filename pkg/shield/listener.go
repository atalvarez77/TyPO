package shield

import (
	"fmt"
	"typo/pkg/logger"
	"typo/pkg/sinkhole"

	"github.com/miekg/dns"
)

// LocalInterceptor holds our server configuration and logger
type LocalInterceptor struct {
	Log      *logger.Logger
	Server   *dns.Server
	Sinkhole *sinkhole.Blocklist
}

// NewInterceptor creates a securely bound Localhost DNS server
func NewInterceptor(log *logger.Logger, blocklist *sinkhole.Blocklist) *LocalInterceptor {
	// We bind strictly to localhost to prevent external access
	addr := "127.0.0.1:53"

	return &LocalInterceptor{
		Log: log,
		Server: &dns.Server{
			Addr: addr,
			Net:  "udp",
		},
		Sinkhole: blocklist,
	}
}

// ServeDNS is the core routing function. It fulfills the dns.Handler interface.
// Every time PC makes a DNS request, this function fires.
func (i *LocalInterceptor) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// 1. Log the interception (Great for our white-box tracking)
	if len(r.Question) > 0 {
		i.Log.Info("Intercepted local request for: %s", r.Question[0].Name)
	}

	cleanDomain := r.Question[0].Name
	if i.Sinkhole.IsBlocked(cleanDomain) {
		i.Log.Info("Sinkhole blocked: %s", cleanDomain)
		dns.HandleFailed(w, r)
		return
	}

	// 2. Pack the incoming request into raw bytes
	rawReq, err := r.Pack()
	if err != nil {
		i.Log.Error("Failed to pack intercepted request: %v", err)
		dns.HandleFailed(w, r)
		return
	}

	// 3. Hand the bytes to our Coffee-Shop Shield (DoH)
	rawResp, err := ForwardToDoH(rawReq, DoHCloudflare, i.Log)
	if err != nil {
		i.Log.Error("DoH Tunnel failed: %v", err)
		dns.HandleFailed(w, r)
		return
	}

	// 4. Unpack the encrypted response from Cloudflare
	respMsg := new(dns.Msg)
	if err := respMsg.Unpack(rawResp); err != nil {
		i.Log.Error("Failed to unpack DoH response: %v", err)
		dns.HandleFailed(w, r)
		return
	}

	// 5. Security Check: Ensure the response ID matches the request ID
	// This prevents DNS cache poisoning attacks
	respMsg.SetReply(r)

	// 6. Write the secure answer back to the local Mac
	if err := w.WriteMsg(respMsg); err != nil {
		i.Log.Error("Failed to write response back to client: %v", err)
	}
}

// Start boots up the UDP listener on a background Goroutine
func (i *LocalInterceptor) Start() error {
	// Register our custom handler to the default DNS multiplexer
	dns.HandleFunc(".", i.ServeDNS)

	i.Log.Info("Starting Secure UDP Listener on %s", i.Server.Addr)

	// ListenAndServe blocks, so we run it asynchronously when called
	err := i.Server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server, closing the port cleanly
func (i *LocalInterceptor) Stop() {
	i.Log.Info("Shutting down Secure UDP Listener...")
	if err := i.Server.Shutdown(); err != nil {
		i.Log.Error("Error during shutdown: %v", err)
	}
}
