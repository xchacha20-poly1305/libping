//go:build !unix

package libping

const isUnix = false

func icmpNetwork(isIPv6 bool) (network string) {
	network = "ip"
	if isIPv6 {
		network += "6:58"
	} else {
		network += "4:1"
	}
	return
}
