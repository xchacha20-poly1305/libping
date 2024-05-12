//go:build !unix

package libping

import (
	"context"
	"runtime"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
)

const isUnix = false

var ErrOS = E.New("not support for: ", runtime.GOOS)

// IcmpPing used to take icmp ping.
// address must be a pure IP address. payload for send.
// If failed, it will return -1, err.
func IcmpPing(ctx context.Context, addr M.Socksaddr, payload []byte) (time.Duration, error) {
	return -1, ErrOS
}
