// Package netguard keeps admin-configured upstreams from reaching sensitive
// network locations (cloud metadata, link-local, unspecified and, unless
// allowed, loopback addresses).
package netguard

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
)

// AllowLoopback permits upstreams on 127.0.0.0/8 and ::1 (local development).
var AllowLoopback = false

var blockedHosts = map[string]bool{
	"metadata.google.internal": true,
	"metadata":                 true,
}

// BlockedIP reports whether ip must not be used as an upstream.
func BlockedIP(ip net.IP) bool {
	switch {
	case ip.IsUnspecified(), ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(), ip.IsMulticast(), ip.IsInterfaceLocalMulticast():
		return true
	case ip.IsLoopback():
		return !AllowLoopback
	}
	return false
}

// CheckHost validates a host (name or IP literal) and every address it resolves to.
func CheckHost(host string) error {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return fmt.Errorf("host is empty")
	}
	if blockedHosts[host] {
		return fmt.Errorf("host %q is not allowed", host)
	}
	if ip := net.ParseIP(host); ip != nil {
		if BlockedIP(ip) {
			return fmt.Errorf("address %s is not allowed", host)
		}
		return nil
	}
	if host == "localhost" && !AllowLoopback {
		return fmt.Errorf("host %q is not allowed", host)
	}
	// Names that don't resolve now may resolve later; the dialer re-checks on connect.
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	for _, ip := range ips {
		if BlockedIP(ip) {
			return fmt.Errorf("host %q resolves to disallowed address %s", host, ip)
		}
	}
	return nil
}

// Dialer returns a dialer that refuses to connect to blocked addresses. The
// check runs on the resolved IP at connect time, which defeats DNS rebinding.
func Dialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			if ip := net.ParseIP(host); ip != nil && BlockedIP(ip) {
				return fmt.Errorf("connection to %s blocked", ip)
			}
			return nil
		},
	}
}

// DialContext adapts Dialer for http.Transport and grpc.WithContextDialer.
func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return Dialer().DialContext(ctx, network, addr)
}
