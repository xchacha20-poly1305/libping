package libping

import (
	"context"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

const DefaultTimeout = 5000 * time.Millisecond

var dialerPool = sync.Pool{
	New: func() any {
		return &net.Dialer{}
	},
}

// TcpPing use TCP to probe `addr`.
func TcpPing(ctx context.Context, controlFunc control.Func, addr M.Socksaddr) (latency time.Duration, err error) {
	dialer := dialerPool.Get().(*net.Dialer)
	defer func() {
		dialer.Control = nil
		dialer.ControlContext = nil
		dialerPool.Put(dialer)
	}()
	dialer.Control = controlFunc

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
