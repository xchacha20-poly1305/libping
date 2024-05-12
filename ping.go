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

const MaxTimeout = 5000 * time.Millisecond

func TcpPing(ctx context.Context, addr M.Socksaddr) (latency time.Duration, err error) {
	dialer := &net.Dialer{}

	if isUnix {
		dialer.ControlContext = func(_ context.Context, network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if FdControl != nil {
					FdControl(int(fd))
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
