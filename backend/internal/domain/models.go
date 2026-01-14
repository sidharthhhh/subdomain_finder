package domain

import "time"

// ResultType indicates the source of the subdomain
type ResultType string

const (
	ResultTypeCT       ResultType = "CT_LOG"
	ResultTypeBrute    ResultType = "BRUTE_FORCE"
	ResultTypeWildcard ResultType = "WILDCARD"
)

// Subdomain represents a discovered subdomain
type Subdomain struct {
	FullDomain string     `json:"domain"`
	Source     ResultType `json:"source"`
	IP         string     `json:"ip,omitempty"`
	FoundAt    time.Time  `json:"found_at"`
}

// Config holds the scan configuration
type Config struct {
	Domain      string
	Wordlist    []string
	Concurrency int
	Timeout     time.Duration
}
