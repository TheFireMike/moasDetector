package parser

import (
	"github.com/pkg/errors"
	"net"
)

func filterPrefix(prefix net.IPNet) error {
	// filter submasks with zero size
	if size, _ := prefix.Mask.Size(); size == 0 {
		return errors.New("prefix size of 0")
	}

	if prefix.IP.To4() != nil {
		// https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
		if prefix.IP[0] == 0 || // "This network"
			prefix.IP.IsPrivate() ||
			(prefix.IP[0] == 100 && prefix.IP[1]&0xC0 == 64) || // Shared Address Space
			prefix.IP.IsLoopback() ||
			prefix.IP.IsLinkLocalUnicast() ||
			(prefix.IP[0] == 192 && prefix.IP[1] == 0 && prefix.IP[2] == 0 && // IETF Protocol Assignments
				prefix.IP[3] != 9 && // Port Control Protocol Anycast
				prefix.IP[3] != 10) || // Traversal Using Relays around NAT Anycast
			(prefix.IP[0] == 192 && prefix.IP[1] == 0 && prefix.IP[2] == 2) || // Documentation (TEST-NET-1)
			(prefix.IP[0] == 198 && prefix.IP[1]&0xFE == 18) || // Benchmarking
			(prefix.IP[0] == 198 && prefix.IP[1] == 51 && prefix.IP[2] == 100) || // Documentation (TEST-NET-2)
			(prefix.IP[0] == 203 && prefix.IP[1] == 0 && prefix.IP[2] == 113) || // Documentation (TEST-NET-3)
			(prefix.IP[0]&0xF0 == 240) || // Reserved
			(prefix.IP.Equal(net.IPv4bcast)) {
			return errors.New("prefix is reserved")
		}
	} else {
		// https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
		if prefix.IP.IsLoopback() ||
			prefix.IP.IsUnspecified() ||
			(prefix.IP[:12].Equal(net.IP{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff})) || // IPv4-mapped Address
			(prefix.IP[:6].Equal(net.IP{0x00, 0x64, 0xff, 0x9b, 0x00, 0x01})) || // IPv4-IPv6 Translat.
			(prefix.IP[:8].Equal(net.IP{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})) || // Discard-Only Address Block
			(prefix.IP[0] == 0x20 && prefix.IP[1] == 0x01 && prefix.IP[2]&0xfE == 0x00 && // IETF Protocol Assignments
				!(prefix.IP[2] == 0x00 && prefix.IP[3] == 0x00) && // TEREDO
				!(prefix.IP.Equal(net.IP{0x20, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})) && // Port Control Protocol Anycast
				!(prefix.IP.Equal(net.IP{0x20, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})) && // Traversal Using Relays around NAT Anycast
				!(prefix.IP[2] == 0x00 && prefix.IP[3] == 0x03) && // AMT
				!(prefix.IP[2] == 0x00 && prefix.IP[3] == 0x04 && prefix.IP[4] == 0x01 && prefix.IP[5] == 0x12) && // AS112-v6
				!(prefix.IP[2] == 0x00 && prefix.IP[3]&0xf0 == 0x20)) || // ORCHIDv2
			(prefix.IP[:4].Equal(net.IP{0x20, 0x01, 0x0d, 0xb8})) || // Documentation
			prefix.IP.IsPrivate() || // Unique-Local
			prefix.IP.IsLinkLocalUnicast() {
			return errors.New("prefix is reserved")
		}
	}

	return nil
}

func filterASN(asn int) error {
	// https://www.iana.org/assignments/iana-as-numbers-special-registry/iana-as-numbers-special-registry.xhtml
	if asn == 0 || // Reserved by [RFC7607]
		asn == 112 || // Used by the AS112 project to sink misdirected DNS queries
		asn == 23456 || // AS_TRANS
		(asn >= 64496 && asn <= 64511) || // For documentation and sample code
		(asn >= 64512 && asn <= 65534) || // For private use
		asn == 65535 || // Reserved by [RFC7300]
		(asn >= 65536 && asn <= 65551) || // For documentation and sample code
		(asn >= 4200000000 && asn <= 4294967294) || // For private use
		asn == 4294967295 { // Reserved by [RFC7300]
		return errors.New("ASN is not public")
	}

	return nil
}
