package libping

import (
	"context"
	"net"
	"syscall"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const MaxTimeout = 5000 * time.Millisecond

// FdControl do some control before connect.
var FdControl func(fd int) = nil

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

// IcmpPing used to take icmp ping.
// If failed, it will return -1, err.
func IcmpPing(ctx context.Context, addr M.Socksaddr, payload []byte) (latency time.Duration, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	isIPv6 := addr.IsIPv6()
	var listenAddress string
	if isIPv6 {
		listenAddress = "::"
	} else {
		listenAddress = "0.0.0.0"
	}

	conn, err := icmp.ListenPacket(icmpNetwork(isIPv6), listenAddress)
	if err != nil {
		return -1, E.Cause(err, "listen")
	}
	_ = context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	if isUnix {
		var rawConn syscall.RawConn
		type sysconn interface {
			SyscallConn() (syscall.RawConn, error)
		}

		if isIPv6 {
			rawConn, err = conn.IPv6PacketConn().PacketConn.(sysconn).SyscallConn()
		} else {
			rawConn, err = conn.IPv4PacketConn().PacketConn.(sysconn).SyscallConn()
		}
		if err != nil {
			return -1, E.Cause(err, "get syscall conn")
		}

		_ = rawConn.Control(func(fd uintptr) {
			if FdControl != nil {
				FdControl(int(fd))
			}
		})
	}

	start := time.Now()

	msg := icmp.Message{
		Body: &icmp.Echo{
			ID:   0xDBB,
			Seq:  0,
			Data: payload,
		},
	}
	if isIPv6 {
		msg.Type = ipv6.ICMPTypeEchoRequest
	} else {
		msg.Type = ipv4.ICMPTypeEcho
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return -1, E.Cause(err, "make icmp message")
	}

	_, err = conn.WriteTo(data, addr.UDPAddr())
	if err != nil {
		return -1, E.Cause(err, "write head")
	}

	_, _, err = conn.ReadFrom(data)
	if err != nil {
		return -1, E.Cause(err, "ReadFrom")
	}

	return time.Since(start), nil
}
