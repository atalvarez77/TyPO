package sinkhole

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const StevenBlackURL = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
const LocalCacheFile = "/tmp/typo_sinkhole_cache.txt"

type Blocklist map[string]bool

func NewBlocklist() Blocklist {
	return make(Blocklist)
}

func UpdateCache() error {
	resp, err := http.Get(StevenBlackURL)
	if err != nil {
		return fmt.Errorf("failed to fetch list: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(LocalCacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (b Blocklist) Load() error {
	file, err := os.Open(LocalCacheFile)
	if err != nil {
		return fmt.Errorf("cache file missing: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == "0.0.0.0" {
			domain := parts[1]
			b[domain] = true
		}
	}
	return scanner.Err()
}

func (b Blocklist) IsBlocked(domain string) bool {
	cleanDomain := strings.TrimSuffix(domain, ".")
	return b[cleanDomain]
}
