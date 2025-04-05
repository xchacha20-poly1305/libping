package libping

import (
	"context"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

const DefaultTimeout = 5000 * time.Millisecond

// TcpPing use TCP to probe `addr`.
func TcpPing(ctx context.Context, addr M.Socksaddr, controlFunc control.Func) (latency time.Duration, err error) {
	dialer := &net.Dialer{
		Control: controlFunc,
	}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, N.NetworkTCP, addr.String())
	if err != nil {
		return -1, E.Cause(err, "dial")
	}
	defer conn.Close()

	return time.Since(start), nil
}

var _ syscall.RawConn = fdProvider(0)

type fdProvider int

func (f fdProvider) Control(ctl func(fd uintptr)) error {
	if ctl == nil {
		return os.ErrInvalid
	}
	ctl(uintptr(f))
	return nil
}

func (f fdProvider) Read(_ func(fd uintptr) (done bool)) error {
	return os.ErrInvalid
}

func (f fdProvider) Write(_ func(fd uintptr) (done bool)) error {
	return os.ErrInvalid
}
