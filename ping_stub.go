//go:build !unix

package libping

import (
	"net"
	"runtime"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
)

var ErrOS = E.New("not support for: ", runtime.GOOS)

// IcmpPing used to take icmp ping.
// address must be a pure IP address. payload for send.
// If failed, it will returns -1, err.
func IcmpPing(address string, timeout time.Duration, payload []byte) (time.Duration, error) {
	return -1, ErrOS
}

func TcpPing(address, port string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(address, port), timeout)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	return time.Since(start), nil
}
