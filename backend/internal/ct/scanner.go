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
			Timeout: 45 * time.Second, // Increased for large domains
		},
	}
}

type CrtShEntry struct {
	NameValue string `json:"name_value"`
}

func (s *CTScanner) Scan(ctx context.Context, targetDomain string) ([]string, error) {
	// Try crt.sh first
	results, err := s.scanCrtSh(ctx, targetDomain)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// If crt.sh failed or found nothing, try fallback (AnubisDB)
	// We log the error but proceed to fallback
	// In a real app we'd inject logger, but for now we just try next.

	resultsFallback, errFallback := s.scanAnubis(ctx, targetDomain)
	if errFallback != nil {
		// If both fail, return the error from the primary or a combined error
		if err != nil {
			return nil, fmt.Errorf("all providers failed. crt.sh: %v, anubis: %v", err, errFallback)
		}
		return nil, errFallback
	}
	return resultsFallback, nil
}

func (s *CTScanner) scanCrtSh(ctx context.Context, targetDomain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%.%s&output=json", targetDomain)

	var resp *http.Response
	var err error

	// Retry logic for 429/5xx errors
	maxRetries := 3 // Reduced since we have fallback
	baseDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err = s.client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				break
			}
			resp.Body.Close()
			if resp.StatusCode == 429 || resp.StatusCode >= 500 {
				time.Sleep(baseDelay * time.Duration(1<<i))
				continue
			}
			return nil, fmt.Errorf("crt.sh returned status: %d", resp.StatusCode)
		}
		time.Sleep(baseDelay * time.Duration(1<<i))
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh failed after retries: %d", resp.StatusCode)
	}

	var entries []CrtShEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode crt.sh response: %v", err)
	}

	uniqueSubs := make(map[string]struct{})
	for _, entry := range entries {
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

func (s *CTScanner) scanAnubis(ctx context.Context, targetDomain string) ([]string, error) {
	url := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", targetDomain)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SubdomainFinder/1.0)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anubis returned status: %d", resp.StatusCode)
	}

	var found []string
	if err := json.NewDecoder(resp.Body).Decode(&found); err != nil {
		return nil, err
	}

	// Anubis returns just a string array, usually clean
	return found, nil
}
