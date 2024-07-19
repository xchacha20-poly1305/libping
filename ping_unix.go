//go:build unix

package libping

import (
	"context"
	"net"
	"os"
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

// IcmpPing used to take icmp ping.
// If failed, it will return -1, err.
// If bindAddr not nil, this function will try to bind it.
func IcmpPing(ctx context.Context,
	addr M.Socksaddr,
	payload []byte,
	bindAddr unix.Sockaddr) (time.Duration, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	var fd int
	if addr.IsIPv6() {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	} else {
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	}
	if err != nil {
		return -1, E.Cause(err, "create socket")
	}

	f := os.NewFile(uintptr(fd), "dgram")

	if FdControl != nil {
		FdControl(ctx, fd)
	}
	if bindAddr != nil {
		err = unix.Bind(fd, bindAddr)
		if err != nil {
			return -1, E.Cause(err, "bind")
		}
	}

	conn, err := net.FilePacketConn(f)
	if err != nil {
		return -1, E.Cause(err, "create conn")
	}
	context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})

	start := time.Now()

	msg := icmp.Message{
		Body: &icmp.Echo{
			ID:   0xDBB,
			Seq:  0,
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
		return -1, E.Cause(err, "read icmp message")
	}

	return time.Since(start), nil
}
