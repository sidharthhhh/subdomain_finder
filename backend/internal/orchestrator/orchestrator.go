package orchestrator

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"subdomain-finder/internal/domain"
)

// Orchestrator coordinates the subdomain discovery process using pluggable components.
type Orchestrator struct {
	resolver   domain.DNSResolver
	ctScanner  domain.CTScanner
	webScraper domain.WebScraper
	wildcard   domain.WildcardDetector

	// Configuration
	wordlistPath string
	numWorkers   int
	logger       *slog.Logger
}

// Option defines a functional configuration option for the Orchestrator.
type Option func(*Orchestrator)

// WithWordlist sets the path to the brute-force wordlist.
func WithWordlist(path string) Option {
	return func(o *Orchestrator) {
		o.wordlistPath = path
	}
}

// WithWorkers sets the number of concurrent DNS resolution workers.
func WithWorkers(n int) Option {
	return func(o *Orchestrator) {
		if n > 0 {
			o.numWorkers = n
		}
	}
}

// WithLogger sets a custom structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(o *Orchestrator) {
		o.logger = l
	}
}

// New creates a new Orchestrator with the given dependencies and options.
func New(
	resolver domain.DNSResolver,
	ctScanner domain.CTScanner,
	webScraper domain.WebScraper,
	wd domain.WildcardDetector,
	opts ...Option,
) *Orchestrator {
	// Default configuration
	o := &Orchestrator{
		resolver:   resolver,
		ctScanner:  ctScanner,
		webScraper: webScraper,
		wildcard:   wd,

		wordlistPath: "wordlist.txt", // Default safe path
		numWorkers:   20,             // Default concurrency
		logger:       slog.Default(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func (o *Orchestrator) Run(ctx context.Context, target string) ([]domain.Subdomain, error) {
	o.logger.Info("starting scan", "target", target, "workers", o.numWorkers, "wordlist", o.wordlistPath)

	// 1. Check for wildcard
	isWildcard, err := o.wildcard.IsWildcard(ctx, target)
	if err != nil {
		o.logger.Error("wildcard detection failed", "error", err, "target", target)
		// Proceeding anyway, assuming explicit scan might still be useful
	}
	if isWildcard {
		o.logger.Warn("wildcard DNS detected", "target", target)
	}

	results := make(chan domain.Subdomain, 100)
	var wg sync.WaitGroup

	// Phase 1: CT Scan
	wg.Add(1)
	go func() {
		defer wg.Done()
		found, err := o.ctScanner.Scan(ctx, target)
		if err != nil {
			o.logger.Error("ct scan failed", "error", err)
			return
		}
		o.logger.Info("ct scan completed", "count", len(found))
		for _, f := range found {
			results <- domain.Subdomain{
				FullDomain: f,
				Source:     domain.ResultTypeCT,
				FoundAt:    time.Now(),
			}
		}
	}()

	// Phase 2: Web Scraping
	wg.Add(1)
	go func() {
		defer wg.Done()
		found, err := o.webScraper.Scan(ctx, target)
		if err != nil {
			o.logger.Error("web scraper failed", "error", err)
			return
		}
		o.logger.Info("web scraper completed", "count", len(found))
		for _, f := range found {
			results <- domain.Subdomain{
				FullDomain: f,
				Source:     "WEB_SCRAPER",
				FoundAt:    time.Now(),
			}
		}
	}()

	// Phase 3: Dynamic Brute Force (Streaming)
	jobs := make(chan string, 100)
	var workerWg sync.WaitGroup

	// Spawn workers
	for i := 0; i < o.numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for subPrefix := range jobs {
				fullDomain := subPrefix + "." + target

				if ctx.Err() != nil {
					return
				}

				// Verify if it resolves
				ip, err := o.resolver.Resolve(ctx, fullDomain)

				if err == nil {
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
	wg.Add(1)
	go func() {
		defer wg.Done()

		file, err := os.Open(o.wordlistPath)
		if err != nil {
			o.logger.Warn("wordlist not found, skipping brute force", "path", o.wordlistPath)
			close(jobs)
			workerWg.Wait()
			return
		}

		o.logger.Info("started streaming wordlist", "path", o.wordlistPath)
		scanner := bufio.NewScanner(file)
		count := 0
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word != "" {
				select {
				case jobs <- word:
					count++
				case <-ctx.Done():
					file.Close()
					close(jobs)
					workerWg.Wait()
					return
				}
			}
		}
		file.Close()
		close(jobs)

		o.logger.Info("finished streaming wordlist", "words_processed", count)
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
		// Dedup logic (keep existing simple validation logic for consistency)
		if res.Source == domain.ResultTypeCT || res.Source == "WEB_SCRAPER" {
			// Attempt to resolve, but keep the result even if it fails (to show full discovery)
			ip, err := o.resolver.Resolve(ctx, res.FullDomain)
			if err == nil {
				res.IP = ip
			} else {
				// Log why it failed but don't drop it.
				// Many CT results are historical or internal (NXDOMAIN)
				o.logger.Debug("resolution failed (keeping entry)", "domain", res.FullDomain, "reason", err)
			}
			uniqueResults[res.FullDomain] = res
		} else {
			uniqueResults[res.FullDomain] = res
		}
	}

	finalList := make([]domain.Subdomain, 0, len(uniqueResults))
	for _, v := range uniqueResults {
		finalList = append(finalList, v)
	}

	o.logger.Info("scan completed", "total_unique", len(finalList))
	return finalList, nil
}
