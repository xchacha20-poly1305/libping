//go:build !unix

package libping

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
)

var ErrOS = E.Cause(os.ErrInvalid, "not support for: ", runtime.GOOS)

// IcmpPing used to take icmp ping.
// address must be a pure IP address. payload for send.
// If failed, it will return -1, err.
func IcmpPing(ctx context.Context, addr M.Socksaddr, payload []byte, controlFnc control.Func) (time.Duration, error) {
	return -1, ErrOS
}
