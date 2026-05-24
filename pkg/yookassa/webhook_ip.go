package yookassa

import (
	"net"
	"net/netip"
	"strings"
)

var yooKassaAllowedNetworks = mustParseNetworks([]string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.156.11/32",
	"77.75.156.35/32",
	"77.75.154.128/25",
	"2a02:5180::/32",
})

func mustParseNetworks(raw []string) []netip.Prefix {
	networks := make([]netip.Prefix, 0, len(raw))

	for _, item := range raw {
		if !strings.Contains(item, "/") {
			item += "/32"
		}

		prefix, err := netip.ParsePrefix(item)
		if err != nil {
			panic(err)
		}

		networks = append(networks, prefix)
	}

	return networks
}

func IsYooKassaIP(clientIP string) bool {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return false
	}

	host := clientIP
	if parsedHost, _, err := net.SplitHostPort(clientIP); err == nil {
		host = parsedHost
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	for _, network := range yooKassaAllowedNetworks {
		if network.Contains(addr) {
			return true
		}
	}

	return false
}
