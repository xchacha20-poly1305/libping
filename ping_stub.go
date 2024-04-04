//go:build !unix

package libping

import (
	"runtime"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
)

var ErrOS = E.New("not support for: ", runtime.GOOS)

// IcmpPing used to take icmp ping.
// address must be a pure IP address. payload for send.
// If failed, it will returns -1, err.
func IcmpPing(address string, timeout time.Duration, payload []byte) (time.Duration, error) {
	return -1, ErrOS
}
