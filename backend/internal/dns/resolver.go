package dns

import (
	"context"
	"fmt"
	"math/rand"
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
			Net:     "tcp", // TCP ensures reliable delivery (no packet loss) at the cost of speed
		},
		// Public DNS servers: Cloudflare, Google, Quad9, OpenDNS
		nameservers: []string{
			"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53",
			"208.67.222.222:53", "8.8.4.4:53", "1.0.0.1:53",
		},
		timeout: timeout,
	}
}

// Errors
var (
	ErrNotFound = fmt.Errorf("domain not found")
	ErrTimeout  = fmt.Errorf("resolution timed out")
)

// Resolve checks if a domain exists by querying A or CNAME records
func (r *Resolver) Resolve(ctx context.Context, domain string) (string, error) {
	// Create a copy of nameservers to shuffle
	servers := make([]string, len(r.nameservers))
	copy(servers, r.nameservers)

	// Fisher-Yates shuffle
	rand.Shuffle(len(servers), func(i, j int) {
		servers[i], servers[j] = servers[j], servers[i]
	})

	var lastErr error
	for _, ns := range servers {
		// Retry up to 3 times per server for robustness
		for attempts := 0; attempts < 3; attempts++ {
			ip, err := r.query(ctx, domain, ns)
			if err == nil {
				return ip, nil
			}
			lastErr = err

			// Optional: slight backoff
			if attempts < 2 {
				time.Sleep(time.Duration(attempts*20) * time.Millisecond)
			}
		}
	}
	return "", lastErr
}

func (r *Resolver) query(ctx context.Context, domain, nameserver string) (string, error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	in, _, err := r.client.ExchangeContext(ctx, m, nameserver)
	if err != nil {
		return "", err
	}

	if in.Rcode != dns.RcodeSuccess {
		return "", ErrNotFound
	}

	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			return a.A.String(), nil
		}
		if cname, ok := answer.(*dns.CNAME); ok {
			return cname.Target, nil
		}
	}

	return "", ErrNotFound
}
