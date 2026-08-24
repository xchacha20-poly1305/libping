package libping

import (
	"context"
	"math"
	"math/rand/v2"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
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

// IcmpPing used to take icmp ping.
// If failed, it will return -1, err.
func IcmpPing(
	ctx context.Context,
	addr M.Socksaddr,
	payload []byte,
	controlFunc control.Func,
) (time.Duration, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	packetConn, err := listenIcmp(ctx, controlFunc, addr)
	if err != nil {
		return -1, E.Cause(err, "listen icmp packet")
	}
	context.AfterFunc(ctx, func() {
		_ = packetConn.Close()
	})
	message, err := buildIcmpMessage(payload, addr.IsIPv6())
	if err != nil {
		return -1, E.Cause(err, "marshall icmp message")
	}
	start := time.Now()
	// Windows: IPAddr Linux: UDPAddr
	_, err = packetConn.WriteTo(message, toNetAddr(addr))
	if err != nil {
		return -1, E.Cause(err, "write packet")
	}
	// Theoretically speaking，request and reply is at the same size.
	buffer := buf.NewSize(len(message))
	defer buffer.Release()
	_, _, err = buffer.ReadPacketFrom(packetConn)
	if err != nil {
		return -1, E.Cause(err, "read icmp message")
	}

	return time.Since(start), nil
}

func buildIcmpMessage(payload []byte, isIPv6 bool) ([]byte, error) {
	message := icmp.Message{
		Body: &icmp.Echo{
			ID:   rand.IntN(math.MaxUint16 + 1),
			Seq:  0,
			Data: payload,
		},
	}
	if isIPv6 {
		message.Type = ipv6.ICMPTypeEchoRequest
	} else {
		message.Type = ipv4.ICMPTypeEcho
	}
	return message.Marshal(nil)
}
