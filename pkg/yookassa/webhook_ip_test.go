package yookassa_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/yookassa"
	"github.com/stretchr/testify/require"
)

func TestIsYooKassaIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientIP string
		want     bool
	}{
		{name: "allowed ipv4", clientIP: "185.71.76.10", want: true},
		{name: "allowed ipv4 with port", clientIP: "185.71.76.10:443", want: true},
		{name: "allowed ipv6", clientIP: "2a02:5180::1", want: true},
		{name: "rejected localhost", clientIP: "127.0.0.1", want: false},
		{name: "rejected empty", clientIP: "", want: false},
		{name: "rejected whitespace", clientIP: "   ", want: false},
		{name: "rejected invalid", clientIP: "not-an-ip", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, yookassa.IsYooKassaIP(tt.clientIP))
		})
	}
}
