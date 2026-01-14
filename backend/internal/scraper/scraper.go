package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type WebScraper struct {
	client *http.Client
}

func New() *WebScraper {
	return &WebScraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *WebScraper) Scan(ctx context.Context, domain string) ([]string, error) {
	// URLs to check: http://domain, https://domain, http://www.domain, https://www.domain
	urls := []string{
		fmt.Sprintf("http://%s", domain),
		fmt.Sprintf("https://%s", domain),
		fmt.Sprintf("http://www.%s", domain),
		fmt.Sprintf("https://www.%s", domain),
	}

	foundMap := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
			if err != nil {
				return
			}
			// Mimic a browser
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

			resp, err := s.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			body := string(bodyBytes)

			// Simple regex to find subdomains
			// Matches: something.domain.tld
			// We need to be careful not to match just "domain.tld" repeatedly or invalid precursors
			// Pattern: [a-zA-Z0-9.-]+\.target\.com
			escapedDomain := regexp.QuoteMeta(domain)
			pattern := fmt.Sprintf(`[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.%s`, escapedDomain)

			re := regexp.MustCompile(pattern)
			matches := re.FindAllString(body, -1)

			mu.Lock()
			for _, m := range matches {
				// Clean up
				m = strings.ToLower(m)
				// Remove leading dots if any
				m = strings.TrimPrefix(m, ".")

				if m != domain && !foundMap[m] {
					foundMap[m] = true
				}
			}
			mu.Unlock()
		}(u)
	}

	wg.Wait()

	results := make([]string, 0, len(foundMap))
	for k := range foundMap {
		results = append(results, k)
	}

	return results, nil
}
