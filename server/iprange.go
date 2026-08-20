package server

import (
	"fmt"
	"net"
	"strings"
)

// ipRange is a comma-separated list of CIDR networks used to authorize
// incoming requests. It is a small, dependency-free replacement for
// github.com/pocke/go-iprange built on top of the standard net/netip package.
type ipRange struct {
	nets []net.IPNet
}

// newIPRange parses allow, a comma-separated list of IPs/CIDRs.
// A bare IPv4 address is treated as /32, a bare IPv6 address as /128.
func newIPRange(allow string) (*ipRange, error) {
	r := &ipRange{}
	for _, raw := range strings.Split(allow, ",") {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if !strings.Contains(s, "/") {
			if strings.Contains(s, ".") {
				s += "/32"
			} else {
				s += "/128"
			}
		}
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", raw, err)
		}
		r.nets = append(r.nets, *n)
	}
	if len(r.nets) == 0 {
		return nil, fmt.Errorf("empty allow list")
	}
	return r, nil
}

// include reports whether addr falls in any of the allowed networks.
func (r *ipRange) include(addr net.IP) bool {
	for _, n := range r.nets {
		if n.Contains(addr) {
			return true
		}
	}
	return false
}

// includeStr is a convenience wrapper around include that parses a string IP.
func (r *ipRange) includeStr(addr string) bool {
	return r.include(net.ParseIP(addr))
}
