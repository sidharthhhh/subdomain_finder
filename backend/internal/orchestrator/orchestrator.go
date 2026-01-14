package orchestrator

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"subdomain-finder/internal/ct"
	"subdomain-finder/internal/dns"
	"subdomain-finder/internal/domain"
	"subdomain-finder/internal/scraper"
	"subdomain-finder/internal/wildcard"
)

type Orchestrator struct {
	resolver     *dns.Resolver
	ctScanner    *ct.CTScanner
	webScraper   *scraper.WebScraper
	wildcard     *wildcard.Detector
	wordlistPath string
}

func New(resolver *dns.Resolver, ctScanner *ct.CTScanner, webScraper *scraper.WebScraper, wd *wildcard.Detector, wordlistPath string) *Orchestrator {
	return &Orchestrator{
		resolver:     resolver,
		ctScanner:    ctScanner,
		webScraper:   webScraper,
		wildcard:     wd,
		wordlistPath: wordlistPath,
	}
}

func (o *Orchestrator) Run(ctx context.Context, target string) ([]domain.Subdomain, error) {
	// 1. Check for wildcard
	isWildcard, _ := o.wildcard.IsWildcard(ctx, target)

	results := make(chan domain.Subdomain, 100)
	var wg sync.WaitGroup

	// Phase 1: CT Scan
	wg.Add(1)
	go func() {
		defer wg.Done()
		found, err := o.ctScanner.Scan(ctx, target)
		if err == nil {
			for _, f := range found {
				results <- domain.Subdomain{
					FullDomain: f,
					Source:     domain.ResultTypeCT,
					FoundAt:    time.Now(),
				}
			}
		}
	}()

	// Phase 2: Web Scraping
	wg.Add(1)
	go func() {
		defer wg.Done()
		found, err := o.webScraper.Scan(ctx, target)
		if err == nil {
			for _, f := range found {
				results <- domain.Subdomain{
					FullDomain: f,
					Source:     "WEB_SCRAPER",
					FoundAt:    time.Now(),
				}
			}
		}
	}()

	// Phase 3: Dynamic Brute Force (Streaming)
	jobs := make(chan string, 100)
	var workerWg sync.WaitGroup

	// Spawn workers
	numWorkers := 20
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for subPrefix := range jobs {
				fullDomain := subPrefix + "." + target

				if ctx.Err() != nil {
					return
				}

				// Verify if it resolves
				ip, found := o.resolver.Resolve(ctx, fullDomain)

				if found {
					if !isWildcard {
						results <- domain.Subdomain{
							FullDomain: fullDomain,
							Source:     domain.ResultTypeBrute,
							IP:         ip,
							FoundAt:    time.Now(),
						}
					}
				}
			}
		}()
	}

	// File Reader Goroutine
	// Independent of 'wg' (which tracks result producers), but we must ensure workers finish.
	// We'll manage worker wait here.
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Open file
		file, err := os.Open(o.wordlistPath)
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				word := strings.TrimSpace(scanner.Text())
				if word != "" {
					select {
					case jobs <- word:
					case <-ctx.Done():
						file.Close()
						close(jobs)
						workerWg.Wait() // Wait for workers to clean up
						return
					}
				}
			}
			file.Close()
		}

		// Done reading or failed to open -> close jobs
		close(jobs)

		// Wait for workers to drain jobs
		workerWg.Wait()
	}()

	// Wait for all phases to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and deduplicate
	uniqueResults := make(map[string]domain.Subdomain)
	for res := range results {
		// Verify results if needed (already verified by DNS worker or coming from trusted sources)
		// Re-verify CT/Scraper just in case? For speed, we previously did. Let's keep it consistent.
		if res.Source == domain.ResultTypeCT || res.Source == "WEB_SCRAPER" {
			if ip, found := o.resolver.Resolve(ctx, res.FullDomain); found {
				res.IP = ip
				uniqueResults[res.FullDomain] = res
			}
		} else {
			uniqueResults[res.FullDomain] = res
		}
	}

	finalList := make([]domain.Subdomain, 0, len(uniqueResults))
	for _, v := range uniqueResults {
		finalList = append(finalList, v)
	}

	return finalList, nil
}
