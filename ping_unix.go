//go:build unix

package libping

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

const isUnix = true

// Do something for fd
var FdControl func(fd int) = nil

// IcmpPing used to take icmp ping.
// If failed, it will return -1, err.
func IcmpPing(ctx context.Context, addr M.Socksaddr, payload []byte) (time.Duration, error) {
	var err error
	var fd int
	if addr.IsIPv6() {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	} else {
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	}

	f := os.NewFile(uintptr(fd), "dgram")
	if err != nil {
		return -1, E.Cause(err, "create file from fd")
	}

	if FdControl != nil {
		FdControl(fd)
	}

	conn, err := net.FilePacketConn(f)
	if err != nil {
		return -1, E.Cause(err, "create conn")
	}

	defer conn.Close()

	start := time.Now()
	timeout := MaxTimeout
	for seq := 1; timeout > 0; seq++ {
		select {
		case <-ctx.Done():
			return -1, E.New(ctx.Err())
		default:
		}

		var sockTo time.Duration
		if timeout > MaxTimeout {
			sockTo = MaxTimeout
		} else {
			sockTo = timeout
		}
		timeout -= sockTo

		err := conn.SetReadDeadline(time.Now().Add(sockTo))
		if err != nil {
			return -1, E.Cause(err, "set read timeout")
		}

		msg := icmp.Message{
			Body: &icmp.Echo{
				ID:   0xDBB,
				Seq:  seq,
				Data: payload,
			},
		}
		if addr.IsIPv6() {
			msg.Type = ipv6.ICMPTypeEchoRequest
		} else {
			msg.Type = ipv4.ICMPTypeEcho
		}

		data, err := msg.Marshal(nil)
		if err != nil {
			return -1, E.Cause(err, "make icmp message")
		}

		err = common.Error(conn.WriteTo(data, &net.UDPAddr{
			IP:   addr.IPAddr().IP,
			Port: 0,
		}))
		if err != nil {
			return -1, E.Cause(err, "write icmp message")
		}

		_, _, err = conn.ReadFrom(data)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				continue
			}

			return -1, E.Cause(err, "read icmp message")
		}

		return time.Since(start), nil
	}

	return -1, E.New("IcmpPing timeout")
}
