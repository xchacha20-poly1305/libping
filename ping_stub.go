//go:build !unix

package libping

const isUnix = false

func icmpNetwork(isIPv6 bool) (network string) {
	network = "ip"
	if isIPv6 {
		network += "6:ipv6-icmp"
	} else {
		network += "4:icmp"
	}
	return
}
