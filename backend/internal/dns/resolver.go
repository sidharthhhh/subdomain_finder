package dns

import (
	"context"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	client      *dns.Client
	nameservers []string
	timeout     time.Duration
}

func NewResolver(timeout time.Duration) *Resolver {
	return &Resolver{
		client: &dns.Client{
			Timeout: timeout,
			Net:     "udp",
		},
		// Public DNS servers: Cloudflare, Google, Quad9
		nameservers: []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"},
		timeout:     timeout,
	}
}

// Resolve checks if a domain exists by querying A or CNAME records
func (r *Resolver) Resolve(ctx context.Context, domain string) (string, bool) {
	// Simple random load balancing is sufficient for this MVP
	// For production, we'd cycle through them or measure latency
	for _, ns := range r.nameservers {
		ip, found := r.query(ctx, domain, ns)
		if found {
			return ip, true
		}
	}
	return "", false
}

func (r *Resolver) query(ctx context.Context, domain, nameserver string) (string, bool) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	// Handle context cancellation / timeout
	// dns.Client.ExchangeContext is available in newer versions,
	// but miekg/dns Exchange doesn't take context directly in older ones.
	// We'll use ExchangeContext if available or manage with goroutines.
	// Checking godoc, ExchangeContext exists.

	in, _, err := r.client.ExchangeContext(ctx, m, nameserver)
	if err != nil {
		return "", false
	}

	if in.Rcode != dns.RcodeSuccess {
		return "", false
	}

	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			return a.A.String(), true
		}
		if cname, ok := answer.(*dns.CNAME); ok {
			// recursively resolve CNAME if needed, or just return success
			// returning CNAME target as "IP" placeholder or resolving it?
			// For subdomain discovery, just existence is usually enough.
			// Let's return the CNAME target.
			return cname.Target, true
		}
	}

	return "", false
}
