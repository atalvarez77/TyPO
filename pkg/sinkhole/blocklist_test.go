package sinkhole

import (
	"os"
	"testing"
)

func TestBlocklistParsing(t *testing.T) {
	mockData := `# Mock StevenBlack Hosts File
127.0.0.1 localhost
127.0.1.1 your-macbook.local

# Blocked malware and ad servers
0.0.0.0 evil-tracker.com
0.0.0.0 annoying-ads.net
`

	err := os.WriteFile(LocalCacheFile, []byte(mockData), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock cache file: %v", err)
	}
	defer os.Remove(LocalCacheFile)

	b := NewBlocklist()
	err = b.Load()
	if err != nil {
		t.Fatalf("Failed to load blocklist: %v", err)
	}

	if !b.IsBlocked("evil-tracker.com.") {
		t.Errorf("Parser failed: evil-tracker.com should be blocked")
	}

	if b.IsBlocked("localhost.") {
		t.Errorf("Parser failed: localhost should remain unblocked")
	}

	if b.IsBlocked("safe-website.com.") {
		t.Errorf("Parser failed: Unlisted domains should remain unblocked")
	}
}
