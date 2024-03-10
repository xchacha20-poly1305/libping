package libping_test

import (
	"testing"

	"github.com/xchacha20-poly1305/libping"
)

const (
	testIPv4Address = "8.8.8.8"
	testIPv6Address = "2001:4860:4860::8888"

	testTimeout int32 = 5000
)

func TestIcmpPing4(t *testing.T) {
	delay, err := libping.IcmpPing(testIPv4Address, testTimeout)
	if err != nil {
		t.Errorf("Ping IPv4 %s: %v", testIPv4Address, err)
		return
	}

	t.Logf("Ping to %s successful. Delay: %d ms", testIPv4Address, delay)
}

func TestIcmpPing6(t *testing.T) {
	delay, err := libping.IcmpPing(testIPv6Address, testTimeout)
	if err != nil {
		t.Errorf("Ping IPv6 %s: %v", testIPv6Address, err)
		return
	}

	t.Logf("Ping to %s successful. Delay: %d ms", testIPv6Address, delay)
}
