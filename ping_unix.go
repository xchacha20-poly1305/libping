//go:build unix

package libping

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/sys/unix"
)

// listenIcmp also returns the underlying *os.File as closer,
// because net.FilePacketConn duplicates the fd and the caller must close both independently.
func listenIcmp(ctx context.Context, controlFunc control.Func, addr M.Socksaddr) (net.PacketConn, io.Closer, error) {
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
		return nil, nil, E.Cause(err, "create socket")
	}
	file := os.NewFile(uintptr(fd), "dgram")
	if file == nil {
		return nil, nil, E.New("failed to create file from fd")
	}
	if controlFunc != nil {
		err = controlFunc(N.NetworkICMP, addr.String(), fdProvider(fd))
		if err != nil {
			_ = file.Close()
			return nil, nil, E.Cause(err, "control")
		}
	}
	packetConn, err := net.FilePacketConn(file)
	if err != nil {
		_ = file.Close()
		return nil, nil, E.Cause(err, "create packet connection")
	}
	return packetConn, file, nil
}

func toNetAddr(addr M.Socksaddr) net.Addr {
	return addr.UDPAddr()
}

// replyBufferSize returns the buffer size needed to read an ICMP reply.
// On Unix, SOCK_DGRAM ICMP sockets strip the IP header before delivery.
func replyBufferSize(msgLen int, isIPv6 bool) int {
	return msgLen
}
