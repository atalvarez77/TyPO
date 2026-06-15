package main

import (
	"fmt"
	"os"
	"time"

	"typo/pkg/dashboard"
	"typo/pkg/ghost"
	"typo/pkg/logger"
	"typo/pkg/shield"
	"typo/pkg/sinkhole"
	"typo/pkg/tunnel"
	"typo/pkg/warden"

	"github.com/getlantern/systray"
)

// Global state tracking
var (
	log              *logger.Logger
	shieldOn         bool
	dnsSrv           *shield.LocalInterceptor
	wardenActive     bool
	portalSuspension bool
	tunnelActive     bool
	vpnClient        tunnel.Provider
)

func init() {
	// Initialize our custom logger
	log = logger.New(logger.LevelDebug)
}

func main() {
	// Security Check: Enforce sudo for Port 53 binding
	if os.Geteuid() != 0 {
		log.Error("FATAL: TyPO must be run with Administrator privileges (sudo) to bind to Port 53.")
		os.Exit(1)
	}

	log.Info("Booting TyPO (Ty Protection Oracle)...")
	systray.Run(onReady, onExit)
}

func onReady() {
	log.Info("Constructing UI components...")
	dashboard.StartServer()
	systray.SetTitle("TyPO")
	systray.SetTooltip("Ty Protection Oracle")

	mShield := systray.AddMenuItem("Coffee-Shop Shield: OFF", "Toggle encrypted DoH interception")
	mUpdateList := systray.AddMenuItem("Update Blocklist", "Download latest StevenBlack hosts")
	mDashboard := systray.AddMenuItem("Open Dashboard Console", "Launch the web-based visual monitor")

	vpnClient = tunnel.NewWarpClient()
	systray.AddSeparator()
	mTunnel := systray.AddMenuItem("VPN Tunnel: OFF", "Toggle full Layer 3 network encapsulation")

	systray.AddSeparator()
	mGhost := systray.AddMenuItem("Ghost Mode: Randomize MAC", "Spoof hardware address for public Wi-Fi")
	mWarden := systray.AddMenuItem("Warden Auto-Detect: ON", "Auto-suspend Shield for login pages")
	wardenActive = true
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit TyPO", "Terminate the application safely")

	blocklist := sinkhole.NewBlocklist()
	dnsSrv = shield.NewInterceptor(log, &blocklist)
	if err := dnsSrv.Sinkhole.Load(); err == nil {
		log.Info("Successfully loaded cached blocklist from disk into memory.")
	} else {
		log.Info("No cached blocklist found. Click 'Update Blocklist' to download one.")
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				// Only check the portal if the Shield and Warden are both active
				if !wardenActive || !shieldOn {
					continue
				}

				trapped := warden.IsTrapped()

				// State 1: We just connected to McDonald's and hit the portal
				if trapped && !portalSuspension {
					log.Info("WARDEN: Captive portal detected. Suspending Shield...")
					shield.RestoreOriginalDNS()
					mShield.SetTitle("Coffee-Shop Shield: SUSPENDED (Portal)")
					portalSuspension = true
				}

				// State 2: You clicked "Accept Terms" and gained internet access
				if !trapped && portalSuspension {
					log.Info("WARDEN: Internet access confirmed. Reactivating Shield...")
					shield.RouteToTyPO()
					mShield.SetTitle("Coffee-Shop Shield: ON")
					portalSuspension = false
				}
			case <-mDashboard.ClickedCh:
				log.Info("Launching visual Dashboard console...")
				dashboard.OpenBrowser()
			case <-mWarden.ClickedCh:
				if wardenActive {
					wardenActive = false
					mWarden.SetTitle("Warden Auto-Detect: OFF")
					log.Info("Warden disabled.")
				} else {
					wardenActive = true
					mWarden.SetTitle("Warden Auto-Detect: ON")
					log.Info("Warden enabled.")
				}
			case <-mShield.ClickedCh:
				if !shieldOn {
					// 1. Turn Shield ON
					log.Info("User requested Shield activation...")

					// Start the local UDP server
					go func() {
						if err := dnsSrv.Start(); err != nil {
							log.Error("Shield failed to start: %v", err)
						}
					}()

					// Update macOS System Routing
					log.Info("Routing macOS traffic to TyPO...")
					if err := shield.RouteToTyPO(); err != nil {
						log.Error("Routing failed: %v", err)
					}

					log.Info("Connected to TyPO. Shield ON.")
					mShield.SetTitle("Coffee-Shop Shield: ON")
					shieldOn = true
				} else {
					// 2. Turn Shield OFF
					log.Info("User requested Shield deactivation...")

					// Stop the local UDP server
					dnsSrv.Stop()

					// Restore macOS System Routing
					log.Info("Restoring macOS default network routing...")
					if err := shield.RestoreOriginalDNS(); err != nil {
						log.Error("Restore failed: %v", err)
					}

					log.Info("Restored macOS default network. Shield OFF.")
					mShield.SetTitle("Coffee-Shop Shield: OFF")
					shieldOn = false
				}
			case <-mUpdateList.ClickedCh:
				log.Info("Downloading StevenBlack blocklist...")
				if err := sinkhole.UpdateCache(); err != nil {
					log.Error("Failed to update blocklist: %v", err)
				} else {
					log.Info("Blocklist downloaded. Loading into memory...")
					if err := dnsSrv.Sinkhole.Load(); err != nil {
						log.Error("Failed to load blocklist: %v", err)
					} else {
						log.Info("Sinkhole active with latest domains.")
					}
				}
			case <-mTunnel.ClickedCh:
				if !tunnelActive {
					log.Info("Activating VPN Tunnel (Layer 3)...")
					if err := vpnClient.Connect(); err != nil {
						log.Error("Failed to engage VPN Tunnel: %v", err)
					} else {
						log.Info("VPN Tunnel active. Traffic is fully encapsulated.")
						mTunnel.SetTitle("VPN Tunnel: ON")
						tunnelActive = true
					}
				} else {
					log.Info("Deactivating VPN Tunnel...")
					if err := vpnClient.Disconnect(); err != nil {
						log.Error("Failed to disconnect VPN Tunnel: %v", err)
					} else {
						log.Info("VPN Tunnel deactivated. Restoring local routing.")
						mTunnel.SetTitle("VPN Tunnel: OFF")
						tunnelActive = false
					}
				}
			case <-mGhost.ClickedCh:
				log.Info("Executing Ghost Mode...")
				newMAC, err := ghost.SpoofMAC()
				if err != nil {
					log.Error("Ghost Mode failed: %v", err)
				} else {
					log.Info("Ghost Mode active. Hardware address spoofed to: %s", newMAC)
					mGhost.SetTitle(fmt.Sprintf("MAC Spoofed: %s", newMAC))
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Info("TyPO shutdown requested...")
	if shieldOn {
		// Restore network settings if the user quits while Shield is ON.
		log.Info("Emergency restoring macOS network routing...")
		shield.RestoreOriginalDNS()
		dnsSrv.Stop()
	}
	if tunnelActive {
		log.Info("Emergency disconnecting VPN Tunnel...")
		if err := vpnClient.Disconnect(); err != nil {
			log.Error("Failed to cleanly disconnect WARP: %v", err)
		}
	}
	log.Info("TyPO shutdown complete. Goodbye.")
}
