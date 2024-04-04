package libping

import (
	"net"
	"os"
	"strings"
	"time"

	E "github.com/sagernet/sing/common/exceptions"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

const payload = "abcdefghijklmnopqrstuvwabcdefghi"

// IcmpPing used to take icmp ping.
// address must be a pure IP address.
// If failed, it will returns -1, err.
func IcmpPing(address string, timeout time.Duration) (time.Duration, error) {
	i := net.ParseIP(address)
	if i == nil {
		return -1, E.New("unable to parse ip ", address)
	}
	var err error
	v6 := i.To4() == nil
	var fd int
	if !v6 {
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	} else {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	}

	f := os.NewFile(uintptr(fd), "dgram")
	if err != nil {
		return -1, E.Cause(err, "create file from fd")
	}

	conn, err := net.FilePacketConn(f)
	if err != nil {
		return -1, E.Cause(err, "create conn")
	}

	defer func(conn net.PacketConn) {
		_ = conn.Close()
	}(conn)

	start := time.Now()
	for seq := 1; timeout > 0; seq++ {
		var sockTo time.Duration
		if timeout > time.Millisecond*1000 {
			sockTo = time.Millisecond * 1000
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
				Data: []byte(payload),
			},
		}
		if !v6 {
			msg.Type = ipv4.ICMPTypeEcho
		} else {
			msg.Type = ipv6.ICMPTypeEchoRequest
		}

		data, err := msg.Marshal(nil)
		if err != nil {
			return -1, E.Cause(err, "make icmp message")
		}

		_, err = conn.WriteTo(data, &net.UDPAddr{
			IP:   i,
			Port: 0,
		})
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

	return 0, E.New("IcmpPing timeout")
}
