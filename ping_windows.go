package libping

import (
	"context"
	"net"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
)

func listenIcmp(ctx context.Context, controlFunc control.Func, addr M.Socksaddr) (net.PacketConn, error) {
	var listenConfig = net.ListenConfig{
		Control: controlFunc,
	}
	var network string
	if addr.IsIPv6() {
		network = "ip6:ipv6-icmp"
	} else {
		network = "ip4:icmp"
	}
	packetConn, err := listenConfig.ListenPacket(ctx, network, "")
	if err != nil {
		return nil, err
	}
	if _, isUDP := packetConn.LocalAddr().(*net.UDPAddr); isUDP {
		_ = packetConn.Close()
		return nil, E.New("listen on UDP because not running in admin")
	}
	return packetConn, nil
}

func toNetAddr(addr M.Socksaddr) net.Addr {
	return addr.IPAddr()
}

// replyBufferSize returns the buffer size needed to read an ICMP reply.
// On Windows, raw ICMP sockets include the IP header in received packets.
// IPv4 header is 20–60 bytes; IPv6 header is 40 bytes fixed.
func replyBufferSize(msgLen int, isIPv6 bool) int {
	var headerSize int
	if isIPv6 {
		headerSize = 40
	} else {
		headerSize = 60
	}
	return msgLen + headerSize
}
