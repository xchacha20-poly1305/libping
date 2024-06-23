//go:build unix

package libping

import (
	N "github.com/sagernet/sing/common/network"
)

const isUnix = true

func icmpNetwork(isIPv6 bool) (network string) {
	network = N.NetworkUDP
	if isIPv6 {
		network += "6"
	} else {
		network += "4"
	}
	return
}
