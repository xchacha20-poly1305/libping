package libping

import (
	"context"
	"net"
	"syscall"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

const DefaultTimeout = 5000 * time.Millisecond

// FdControl do some control before connect.
var FdControl func(ctx context.Context, fd int) = nil

// TcpPing use TCP to probe `addr`.
// In unix, this function will use dialer.ControlContext.
func TcpPing(ctx context.Context, dialer *net.Dialer, addr M.Socksaddr) (latency time.Duration, err error) {
	if isUnix {
		oldControl := dialer.ControlContext
		dialer.ControlContext = func(ctx context.Context, network, address string, c syscall.RawConn) error {
			if oldControl != nil {
				err := oldControl(ctx, network, address, c)
				if err != nil {
					return err
				}
			}
			return c.Control(func(fd uintptr) {
				if FdControl != nil {
					FdControl(ctx, int(fd))
				}
			})
		}
	}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, N.NetworkTCP, addr.String())
	if err != nil {
		return -1, E.Cause(err, "dial")
	}
	defer conn.Close()

	return time.Since(start), nil
}
