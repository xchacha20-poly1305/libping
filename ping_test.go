package libping

import (
	"testing"

	F "github.com/sagernet/sing/common/format"
)

const (
	testIPv4Address = "8.8.8.8"
	testIPv6Address = "2001:4860:4860::8888"
)

var (
	payload = []byte("abcdefghijklmnopqrstuvwxyz")
)

func TestIcmpPing(t *testing.T) {
	tt := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "Domain",
			address: "i.local",
			wantErr: true,
		},
		{
			name:    "IPv4",
			address: testIPv4Address,
			wantErr: false,
		},
		{
			name:    "IPv6",
			address: testIPv6Address,
			wantErr: false,
		},
	}

	for _, test := range tt {
		delay, err := IcmpPing(test.address, MaxTimeout, payload)
		if (err != nil) != test.wantErr {
			t.Errorf("Test %s failed: %v", test.name, err)
			continue
		}

		t.Logf("Test %s successful. Delay: %s", test.name, F.ToString(delay))
	}
}

func TestTcpPing(t *testing.T) {
	tt := []struct {
		name          string
		address, port string
		wantErr       bool
	}{
		{
			name:    "Domain",
			address: "sekai.icu",
			port:    "443",
			wantErr: true,
		},
		{
			name:    "Miss address",
			address: "",
			port:    "443",
			wantErr: true,
		},
		{
			name:    "IPv4",
			address: testIPv4Address,
			port:    "53",
			wantErr: false,
		},
		{
			name:    "IPv6",
			address: testIPv6Address,
			port:    "53",
			wantErr: false,
		},
	}

	for _, test := range tt {
		latency, err := TcpPing(test.address, test.port, MaxTimeout)
		if (err != nil) != test.wantErr {
			t.Errorf("Failed to test %s: %v", test.name, err)
			continue
		}
		t.Logf("Tested %s, latency: %d ms", test.name, latency.Milliseconds())
	}
}
