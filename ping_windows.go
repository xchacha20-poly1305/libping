package libping

import (
	"context"
	"io"
	"net"

	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
)

func listenIcmp(ctx context.Context, controlFunc control.Func, addr M.Socksaddr) (net.PacketConn, io.Closer, error) {
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
		return nil, nil, err
	}
	if _, isUDP := packetConn.LocalAddr().(*net.UDPAddr); isUDP {
		_ = packetConn.Close()
		return nil, nil, E.New("listen on UDP because not running in admin")
	}
	return packetConn, nil, nil
}

func toNetAddr(addr M.Socksaddr) net.Addr {
	return addr.IPAddr()
}
