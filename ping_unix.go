//go:build unix

package libping

import (
	"context"
	"net"
	"os"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/sys/unix"
)

func listenIcmp(ctx context.Context, controlFunc control.Func, addr M.Socksaddr) (net.PacketConn, error) {
	var (
		fd  int
		err error
	)
	if addr.IsIPv6() {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	} else {
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	}
	if err != nil {
		return nil, E.Cause(err, "create socket")
	}

	file := os.NewFile(uintptr(fd), "dgram")

	if controlFunc != nil {
		err = controlFunc(N.NetworkICMP, addr.String(), fdProvider(fd))
		if err != nil {
			return nil, E.Cause(err, "control")
		}
	}

	return net.FilePacketConn(file)
}

func toNetAddr(addr M.Socksaddr) net.Addr {
	return addr.UDPAddr()
}
