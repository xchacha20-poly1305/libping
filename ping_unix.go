//go:build unix

package libping

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

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

	var (
		err error
		fd  int

		isIPv6 = addr.IsIPv6()
	)
	if isIPv6 {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	} else {
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	}
	if err != nil {
		return -1, E.Cause(err, "create socket")
	}

	file := os.NewFile(uintptr(fd), "dgram")

	if controlFunc != nil {
		var network string
		if isIPv6 {
			network = N.NetworkICMPv6
		} else {
			network = N.NetworkICMPv4
		}
		err = controlFunc(network, addr.String(), fdProvider(fd))
		if err != nil {
			return -1, E.Cause(err, "control")
		}
	}

	conn, err := net.FilePacketConn(file)
	if err != nil {
		return -1, E.Cause(err, "create conn")
	}
	_ = context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})

	start := time.Now()

	message := icmp.Message{
		Body: &icmp.Echo{
			ID:   0xDBB,
			Seq:  0,
			Data: payload,
		},
	}
	if isIPv6 {
		message.Type = ipv6.ICMPTypeEchoRequest
	} else {
		message.Type = ipv4.ICMPTypeEcho
	}

	binaryMessage, err := message.Marshal(nil)
	if err != nil {
		return -1, E.Cause(err, "make icmp message")
	}

	_, err = conn.WriteTo(binaryMessage, addr.UDPAddr())
	if err != nil {
		return -1, E.Cause(err, "write icmp message")
	}

	_, _, err = conn.ReadFrom(binaryMessage)
	if err != nil {
		return -1, E.Cause(err, "read icmp message")
	}

	return time.Since(start), nil
}
