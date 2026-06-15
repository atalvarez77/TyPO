🛡️ TyPO (Ty Protection Oracle)
A localized, defense-in-depth network security daemon for macOS.

TyPO is a custom-engineered, lightweight security utility designed to protect network metadata and hardware identity on untrusted public Wi-Fi networks. Operating invisibly from the macOS menu bar, it combines a Layer 7 DNS Sinkhole, DNS-over-HTTPS (DoH) proxy, Layer 2 MAC randomization, and automated captive portal handling into a single, cohesive binary.

📖 The Origin Story
This project was born out of necessity during my five-month study abroad program at Waseda University in Tokyo. Navigating between my apartment in Kita-Ikebukuro, university campus networks, and local cafes (like McDonald's), I realized how severely exposed my MacBook was to local network snooping, recurring device tracking, and aggressive telemetry.

I didn't want to rely on a fragmented collection of heavy commercial VPNs and ad-blockers that drained battery life and required constant manual toggling. I needed a zero-latency, local-first tool that understood exactly when to raise its shields and when to drop them for captive portals.

TyPO is the result of mapping my personal threat model to the OSI stack: neutralizing local Layer 2 tracking and Layer 7 metadata interception while leaving standard web routing fast and uninhibited.

⚙️ Core Architecture & Features
TyPO operates natively on macOS, bridging system-level networking commands with a highly concurrent Go backend.

The Coffee-Shop Shield (DoH Proxy): Intercepts local plaintext DNS requests (Port 53) and double-encapsulates them into TLS 1.3 HTTPS packets (Port 443) routed to Cloudflare (1.1.1.1). This completely blinds local cafe routers and ISPs from seeing your browsing destinations.

Zero-Latency Sinkhole: Before any DNS request leaves the machine, TyPO checks it against an in-memory Hash Map populated by the StevenBlack hosts list. Known tracking, ad, and malware domains are dropped into the void in microseconds.

Ghost Mode (MAC Randomizer): Defends against Layer 2 hardware tracking by injecting a locally-administered, randomized hexadecimal MAC address directly into the Apple Silicon Wi-Fi interface.

The Warden (Captive Portal Auto-Detect): A background Goroutine that silently polls Apple's native connectivity URL (captive.apple.com). When it detects a coffee shop login trap, the Warden automatically suspends the DNS shield to allow authentication, reinstating the lock-down the moment internet access is achieved.

Local Web Dashboard: A decoupled HTTP server running locally on 127.0.0.1:9090, serving a real-time console of network requests and blocked threats without bogging down the macOS UI thread.

🛠️ Tech Stack
Language: Go (Golang) — Chosen for its static compilation, incredible concurrency (Goroutines), and raw network socket performance.

UI Framework: getlantern/systray for the native macOS menu bar integration, paired with standard HTML/JS for the decoupled web dashboard.

System APIs: macOS networksetup and ifconfig command-line utilities.

Deployment: Stripped binary (-s -w) wrapped in an AppleScript/Automator .app bundle, leveraging sudoers for secure, password-less daemon execution.

🚧 Engineering Challenges & Solutions
Building a root-level networking tool on modern macOS (especially Apple Silicon) presented severe architectural hurdles:

Challenge: Apple's Hardware MAC Address Lock.
Modern macOS strictly prohibits altering the physical MAC address if the Wi-Fi card is associated with a network or powered off.
Solution: Engineered a precise Power-Cycle Injection sequence. TyPO forces the interface OFF, turns it back ON, and executes the spoofing command in the ~200ms window before macOS auto-handshakes with saved networks.

Challenge: Main-Thread UI Panics.
Native GUI frameworks heavily conflict with the systray package on macOS, crashing the application when competing for the OS main rendering thread.
Solution: Adopted the "Pi-hole Architecture." TyPO spins up a lightweight background HTTP server and commands the native web browser to open the UI, completely isolating the visual layer from the core daemon.

Challenge: Captive Portal Lockouts.
Public Wi-Fi requires a web login, but the DoH Shield routes DNS via Port 443, bypassing the local router's DNS hijack and permanently locking the user out of the internet.
Solution: Built The Warden, which hijacks macOS's native captive portal detection loop. It acts as an autonomous state machine, dropping local routing to 127.0.0.1 when trapped, and restoring it upon success.

Challenge: Read-Only Daemon Panics.
When packaged as an Automator .app launching on startup, TyPO executes from the macOS root directory (/), triggering read-only file system errors when downloading blocklists.
Solution: Refactored file operations to strictly utilize absolute paths pointing to /tmp/, ensuring the daemon always has secure, writable scratch space regardless of its execution context.

🚀 How to Run It
Because TyPO intercepts Port 53, it must run with administrator (root) privileges. It is designed to be deployed as a set-and-forget background application.

1. Compile the Optimized Engine
Clone the repository and compile the Go code into a stripped, standalone machine-code binary:

go build -ldflags="-s -w" -o typo_engine main.go
sudo mv typo_engine /usr/local/bin/

2. Configure the Sudoers Bypass
To allow TyPO to launch silently on boot without prompting for a password, add an exception to your sudoers file:

sudo visudo -f /private/etc/sudoers.d/typo

Add the following line (replace YOUR_USERNAME with your Mac username):
YOUR_USERNAME ALL=(ALL) NOPASSWD: /usr/local/bin/typo_engine

3. Create the Native macOS Wrapper
Open Automator and create a new Application.
Add a Run AppleScript block with the following code to detach the background process:

on run {input, parameters}
    do shell script "sudo /usr/local/bin/typo_engine > /dev/null 2>&1 &"
    return input
end run

Save as TyPO.app in your /Applications folder.

4. Execution
Double-click TyPO.app or add it to System Settings > Login Items for auto-launch. The TyPO shield icon will silently appear in your menu bar, ready to deploy. Click Open Dashboard Console to monitor real-time network sinkhole statistics.
