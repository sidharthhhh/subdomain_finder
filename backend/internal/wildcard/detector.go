package wildcard

import (
	"context"
	"fmt"
	"math/rand"

	"subdomain-finder/internal/dns"
)

type Detector struct {
	resolver *dns.Resolver
}

func NewDetector(resolver *dns.Resolver) *Detector {
	return &Detector{
		resolver: resolver,
	}
}

// IsWildcard checks if the domain resolves to a generic address for random subdomains
func (d *Detector) IsWildcard(ctx context.Context, domain string) (bool, error) {
	// Generate a random subdomain that definitely shouldn't exist
	randSub := fmt.Sprintf("wildcard-test-%d.%s", rand.Int63(), domain)

	_, err := d.resolver.Resolve(ctx, randSub)
	found := (err == nil)
	return found, nil
}
