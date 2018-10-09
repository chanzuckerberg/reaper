package util

import (
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
)

// ContainsPublicIps is legit
func ContainsPublicIps(cidrblock string) bool {

	_, a, _ := net.ParseCIDR("10.0.0.0/8")
	_, b, _ := net.ParseCIDR("172.16.0.0/12")
	_, c, _ := net.ParseCIDR("192.168.0.0/16")

	_, network, _ := net.ParseCIDR(cidrblock)

	if SubnetContainsRange(a, network) ||
		SubnetContainsRange(b, network) ||
		SubnetContainsRange(c, network) {
		return false
	}
	return true
}

func SubnetContainsRange(haystack, needle *net.IPNet) bool {
	first, last := cidr.AddressRange(needle)
	if haystack.Contains(first) && haystack.Contains(last) {
		return true
	}

	return false
}
