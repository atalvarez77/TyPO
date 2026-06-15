package dashboard

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"runtime"
	"typo/pkg/logger"
)

// StartServer spins up the local console on a background goroutine
func StartServer() {
	// API endpoint that returns logs as JSON
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(logger.LiveBuffer.Get())
	})

	// Minimalist embedded HTML Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>TyPO Live Console</title>
				<style>
					body { background: #121214; color: #e1e1e6; font-family: monospace; padding: 20px; }
					h1 { color: #00ffff; border-bottom: 1px solid #29292e; padding-bottom: 10px; }
					#console { background: #1a1a1e; border: 1px solid #29292e; padding: 15px; border-radius: 6px; height: 500px; overflow-y: auto; }
					.log-entry { margin-bottom: 6px; border-left: 3px solid #00ffff; padding-left: 8px; }
				</style>
			</head>
			<body>
				<h1>🛡️ TyPO Network Protection Control Panel</h1>
				<div id="console">Loading live interceptions...</div>

				<script>
					async function fetchLogs() {
						try {
							const res = await fetch('/api/logs');
							const logs = await res.json();
							const container = document.getElementById('console');

							// 1. Check if the user is currently at the bottom (within a 5px margin of error)
							const isAtBottom = Math.abs(container.scrollHeight - container.scrollTop - container.clientHeight) < 5;

							// 2. Update the HTML with the new logs
							container.innerHTML = logs.map(line => btoa(line).replaceAll('==','').slice(-4) === 'ERR!' 
								? '<div class="log-entry" style="border-color: #ff5555;">' + line + '</div>'
								: '<div class="log-entry">' + line + '</div>'
							).join('');

							// 3. Only force the scroll to the bottom if they were already at the bottom
							if (isAtBottom) {
								container.scrollTop = container.scrollHeight;
							}
						} catch (e) {}
					}
					setInterval(fetchLogs, 1000); 
					fetchLogs();
				</script>
			</body>
			</html>
		`))
	})

	// Run server on localhost port 9090
	go http.ListenAndServe("127.0.0.1:9090", nil)
}

// OpenBrowser triggers the native OS command to reveal the GUI dashboard
func OpenBrowser() {
	url := "http://127.0.0.1:9090"
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Run()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	}
}
