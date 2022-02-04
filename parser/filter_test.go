package parser

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestFilterPrefix_IPv4_ThisNetwork(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("0.2.0.0/16")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("1.0.0.0/8")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_Private(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("10.2.0.0/16")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("172.16.48.0/24")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("192.168.3.0/24")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("11.0.0.0/16")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_SharedAddressSpace(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("100.65.0.0/16")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("100.128.0.0/10")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_Loopback(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("127.1.0.0/16")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("128.0.0.0/8")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_LinkLocal(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("169.254.4.0/24")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("169.253.0.0/16")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_IETFProtocolAssignments(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("192.0.0.128/26")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("192.0.1.0/24")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_PortControlProtocolAnycast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("192.0.0.9/32")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_TraversalUsingRelaysAroundNATAnycast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("192.0.0.10/32")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_NAT64DNS64Discovery(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("192.0.0.170/32")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("192.0.0.171/32")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_DocumentationTESTNET1(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("192.0.2.128/26")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("192.0.3.0/24")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_Benchmarking(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("198.19.0.0/16")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("198.20.0.0/15")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_DocumentationTESTNET2(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("198.51.100.128/26")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("198.51.101.0/24")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_DocumentationTESTNET3(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("203.0.113.128/25")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("203.0.114.0/24")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_Reserved(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("241.0.0.0/8")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("224.0.0.0/4")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv4_LimitedBroadcast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("255.255.255.255/32")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_LoopbackAddress(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("::1/128")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("::2/128")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_UnspecifiedAddress(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("::/128")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_IPv4MappedAddress(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("::ffff:14:0/112")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("::fffe:0:0/96")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_IPv4IPv6Translation(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("64:ff9b:1:42::/64")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("64:ff9b:2::/48")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_DiscardOnlyAddressBlock(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("100:0:0:0:1::/80")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("100:0:0:1::/64")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_IETFProtocolAssignments(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:100::/24")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:200::/23")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_TEREDO(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:0:F000::/36")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:1::/32")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_PortControlProtocolAnycast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:1::1/128")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:1::0/128")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_TraversalUsingRelaysAroundNATAnycast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:1::2/128")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:1::3/128")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_Benchmarking(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:2:0:1000::/54")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_AMT(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:3:F000::/36")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:4::/32")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_AS112v6(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:4:112:F000::/52")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:4:113::/48")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_ORCHIDv2(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:21::/32")
	assert.Nil(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:30::/28")
	assert.Error(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_Documentation(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("2001:db8:F000::/36")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("2001:db9::/32")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_UniqueLocal(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("fd00::/8")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("fe00::/7")
	assert.Nil(t, filterPrefix(*prefix))
}

func TestFilterPrefix_IPv6_LinkLocalUnicast(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("fe90::/12")
	assert.Error(t, filterPrefix(*prefix))

	_, prefix, _ = net.ParseCIDR("fec0::/10")
	assert.Nil(t, filterPrefix(*prefix))
}
