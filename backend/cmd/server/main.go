package main

import (
	"log"
	"net/http"
	"time"

	"subdomain-finder/internal/api"
	"subdomain-finder/internal/ct"
	"subdomain-finder/internal/dns"
	"subdomain-finder/internal/orchestrator"
	"subdomain-finder/internal/scraper"
	"subdomain-finder/internal/wildcard"
)

func main() {
	// Initialize components
	resolver := dns.NewResolver(2 * time.Second)
	scanner := ct.NewScanner()
	webScraper := scraper.New()
	detector := wildcard.NewDetector(resolver)
	orch := orchestrator.New(resolver, scanner, webScraper, detector, "wordlist.txt")
	handler := api.NewHandler(orch)

	// Setup routes
	http.HandleFunc("/scan", handler.HandleScan)

	// Start server
	log.Println("Starting Subdomain Finder API on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
