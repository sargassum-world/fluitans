package client

import (
	"net/netip"

	"github.com/pkg/errors"
)

func SplitIPAddresses(rawAddresses []string) (ipv4 []string, ipv6 []string, err error) {
	ipv4 = make([]string, 0, len(rawAddresses))
	ipv6 = make([]string, 0, len(rawAddresses))
	for _, rawAddress := range rawAddresses {
		address, err := netip.ParseAddr(rawAddress)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "couldn't parse IP address %s", rawAddress)
		}
		if address.Is4() {
			ipv4 = append(ipv4, address.String())
			continue
		}
		if address.Is6() {
			ipv6 = append(ipv6, address.String())
			continue
		}
	}
	return ipv4, ipv6, nil
}
