package domain

import "context"

type DNSResolver interface {
	Resolve(ctx context.Context, domain string) (string, error)
}

type CTScanner interface {
	Scan(ctx context.Context, domain string) ([]string, error)
}

type WebScraper interface {
	Scan(ctx context.Context, domain string) ([]string, error)
}

type WildcardDetector interface {
	IsWildcard(ctx context.Context, domain string) (bool, error)
}
