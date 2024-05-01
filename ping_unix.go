//go:build unix

package libping

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/x/constraints"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

// Do something for fd
var FdControl func(fd int) = nil

// IcmpPing used to take icmp ping.
// address must be a pure IP address. payload for send.
// If failed, it will returns -1, err.
func IcmpPing(address string, timeout time.Duration, payload []byte) (time.Duration, error) {
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

	if FdControl != nil {
		FdControl(fd)
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

func TcpPing(address, port string, timeout time.Duration) (latency time.Duration, err error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return -1, E.New("failed to parse ip: ", address)
	}
	isIPv6 := ip.To4() == nil

	var socketFd int
	if isIPv6 {
		socketFd, err = unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, 0)
	} else {
		socketFd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	}
	if err != nil {
		return -1, err
	}
	defer unix.Close(socketFd)

	var timeval unix.Timeval
	microseconds := timeout.Microseconds()
	castAssignInteger(microseconds/1e6, &timeval.Sec)
	// Specifying the type explicitly is not necessary here, but it makes GoLand happy.
	castAssignInteger[int64](microseconds%1e6, &timeval.Usec)
	_ = unix.SetsockoptTimeval(socketFd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &timeval)

	if FdControl != nil {
		FdControl(socketFd)
	}

	var sockAddr unix.Sockaddr
	portInt, _ := strconv.Atoi(port)
	if isIPv6 {
		sockAddr = &unix.SockaddrInet6{Port: portInt, Addr: [16]byte(ip.To16())}
	} else {
		sockAddr = &unix.SockaddrInet4{Port: portInt, Addr: [4]byte(ip.To4())}
	}

	start := time.Now()
	err = unix.Connect(socketFd, sockAddr)
	if err != nil {
		return -1, err
	}

	return time.Since(start), nil
}

func castAssignInteger[T, R constraints.Integer](from T, to *R) {
	*to = R(from)
}
