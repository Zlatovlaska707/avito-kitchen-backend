package webhook

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

var allowedPorts = map[int]struct{}{
	80: {}, 443: {}, 8080: {}, 8081: {}, 3000: {},
}

type SSRFGuard struct {
	AllowPrivateIPs bool
	Resolver        func(ctx context.Context, host string) ([]net.IP, error)
}

func DefaultResolver(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

func (g SSRFGuard) ValidateURL(ctx context.Context, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: parse url: %w", usecase.ErrWebhookURLForbidden, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return usecase.ErrWebhookURLForbidden
	}

	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasPrefix(host, "metadata.") {
		return usecase.ErrWebhookURLForbidden
	}

	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	var portNum int
	if _, scanErr := fmt.Sscanf(port, "%d", &portNum); scanErr != nil {
		return usecase.ErrWebhookURLForbidden
	}
	if _, ok := allowedPorts[portNum]; !ok {
		return usecase.ErrWebhookURLForbidden
	}

	if g.AllowPrivateIPs {
		return nil
	}

	resolver := g.Resolver
	if resolver == nil {
		resolver = DefaultResolver
	}
	ips, err := resolver(ctx, host)
	if err != nil {
		return fmt.Errorf("%w: resolve host: %w", usecase.ErrWebhookURLForbidden, err)
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return usecase.ErrWebhookURLForbidden
		}
	}
	return nil
}

func (g SSRFGuard) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if err := g.ValidateURL(ctx, "http://"+host+":"+port); err != nil && !g.AllowPrivateIPs {
		return nil, err
	}
	d := &net.Dialer{}
	return d.DialContext(ctx, network, addr)
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	privateRanges := []string{
		"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"169.254.0.0/16", "100.64.0.0/10", "::1/128", "fc00::/7", "fe80::/10",
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
