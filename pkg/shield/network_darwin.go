//go:build darwin

package shield

import (
	"fmt"
	"os/exec"
	"strings"
)

// PrimaryInterface defines the default macOS network service.
// For MacBooks, this is almost always "Wi-Fi".
const PrimaryInterface = "Wi-Fi"

// RouteToTyPO executes macOS system commands to force DNS to our local interceptor
func RouteToTyPO() error {
	// The networksetup command is macOS's native terminal utility for network preferences
	cmd := exec.Command("networksetup", "-setdnsservers", PrimaryInterface, "127.0.0.1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set macOS DNS: %v (Output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// RestoreOriginalDNS clears the custom DNS, returning control to the local Wi-Fi router
func RestoreOriginalDNS() error {
	// Passing "Empty" tells macOS to clear custom DNS and grab the default from DHCP
	cmd := exec.Command("networksetup", "-setdnsservers", PrimaryInterface, "Empty")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restore macOS DNS: %v (Output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}
