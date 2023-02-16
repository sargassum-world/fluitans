package zerotier

import (
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func Get6Plane(networkID, address string) (string, error) {
	// Taken from BSD-3-Clause-licensed project
	// https://github.com/zerotier/terraform-provider-zerotier/blob/main/pkg/zerotier/member.go
	networkIDInt, err := strconv.ParseUint(networkID, 16, 64)
	if err != nil {
		return "", errors.Wrap(err, "couldn't parse zerotier network id")
	}

	const maskOffset = 32
	networkMask := uint32((networkIDInt >> maskOffset) ^ networkIDInt)
	networkPrefix := strconv.FormatUint(uint64(networkMask), 16)
	return net.ParseIP(buildIPv6("fc" + networkPrefix + address + "000000000001")).String(), nil
}

func GetRFC4193(networkID, address string) (string, error) {
	// Taken from BSD-3-Clause-licensed project
	// https://github.com/zerotier/terraform-provider-zerotier/blob/main/pkg/zerotier/member.go
	return net.ParseIP(buildIPv6("fd" + networkID + "9993" + address)).String(), nil
}

func buildIPv6(data string) string {
	// Taken from BSD-3-Clause-licensed project
	// https://github.com/zerotier/terraform-provider-zerotier/blob/main/pkg/zerotier/member.go
	s := strings.SplitAfter(data, "")
	end := len(s) - 1
	result := ""
	for i, s := range s {
		result += s
		if (i+1)%4 == 0 && i != end {
			result += ":"
		}
	}
	return result
}
