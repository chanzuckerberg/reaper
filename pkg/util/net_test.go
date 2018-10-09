package util_test

import (
	"net"
	"testing"

	"github.com/chanzuckerberg/reaper/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestContainsPublicIps(t *testing.T) {
	a := assert.New(t)
	one := "1.1.1.1/0"
	a.True(util.ContainsPublicIps(one))
}

func TestSubnetContainsRange(t *testing.T) {
	a := assert.New(t)

	_, x, _ := net.ParseCIDR("10.0.0.0/8")
	_, y, _ := net.ParseCIDR("10.1.0.0/16")
	a.True(util.SubnetContainsRange(x, y))

	_, x, _ = net.ParseCIDR("10.0.0.0/8")
	_, y, _ = net.ParseCIDR("11.0.0.0/16")
	a.False(util.SubnetContainsRange(x, y))
}
