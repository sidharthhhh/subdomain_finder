package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CTScanner struct {
	client *http.Client
}

func NewScanner() *CTScanner {
	return &CTScanner{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type CrtShEntry struct {
	NameValue string `json:"name_value"`
}

func (s *CTScanner) Scan(ctx context.Context, targetDomain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%.%s&output=json", targetDomain)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned status: %d", resp.StatusCode)
	}

	var entries []CrtShEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		// Sometimes crt.sh returns malformed results or HTML on error
		return nil, fmt.Errorf("failed to decode crt.sh response: %v", err)
	}

	uniqueSubs := make(map[string]struct{})
	for _, entry := range entries {
		// Clean up the domain names (remove wildcards, handle multi-line)
		names := strings.Split(entry.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(name)
			name = strings.TrimPrefix(name, "*.")
			if strings.HasSuffix(name, targetDomain) {
				uniqueSubs[name] = struct{}{}
			}
		}
	}

	var results []string
	for sub := range uniqueSubs {
		results = append(results, sub)
	}

	return results, nil
}
