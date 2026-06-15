//go:build darwin

package ghost

import (
	"crypto/rand"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PrimaryInterface defines the default MacBook Wi-Fi hardware port.
const PrimaryInterface = "en0"

// GenerateRandomMAC creates a valid, locally administered hardware address.
func GenerateRandomMAC() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	// Set the locally administered bit and clear the multicast bit
	buf[0] = (buf[0] | 2) & 0xfe

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}

// SpoofMAC changes the physical hardware address of the Wi-Fi card.
func SpoofMAC() (string, error) {
	newMAC, err := GenerateRandomMAC()
	if err != nil {
		return "", fmt.Errorf("failed to generate MAC: %v", err)
	}

	// 1. Force a network disconnect by powering off the Wi-Fi interface
	exec.Command("networksetup", "-setairportpower", PrimaryInterface, "off").Run()

	// 2. Power the Wi-Fi interface back on so it can accept the hardware change
	exec.Command("networksetup", "-setairportpower", PrimaryInterface, "on").Run()

	// 3. Wait a tiny fraction of a second for the hardware to boot up,
	// but execute BEFORE it can auto-connect to a saved network
	time.Sleep(200 * time.Millisecond)

	// 4. Inject the spoofed MAC address
	cmd := exec.Command("ifconfig", PrimaryInterface, "ether", newMAC)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to change MAC address: %v (Output: %s)", err, strings.TrimSpace(string(output)))
	}

	return newMAC, nil
}
