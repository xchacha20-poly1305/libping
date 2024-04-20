package libping

import (
	"testing"
	"time"

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
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "Domain",
			address: "i.local",
			timeout: MaxTimeout,
			wantErr: true,
		},
		{
			name:    "IPv4",
			address: testIPv4Address,
			timeout: MaxTimeout,
			wantErr: false,
		},
		{
			name:    "IPv6",
			address: testIPv6Address,
			timeout: MaxTimeout,
			wantErr: false,
		},
	}

	for _, test := range tt {
		delay, err := IcmpPing(test.address, test.timeout, payload)
		if (err != nil) != test.wantErr {
			t.Errorf("Test %s failed: %v", test.name, err)
			return
		}

		t.Logf("Test %s successful. Delay: %s", test.name, F.ToString(delay))
	}
}
