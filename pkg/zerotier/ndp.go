package zerotier

import (
	"encoding/hex"
	"net"
	"strings"
)

func Get6Plane(networkID, address string) (string, error) {
	idBytes, err := hex.DecodeString(networkID)
	if err != nil {
		return "", err
	}

	for i, b := range idBytes {
		if i+4 < len(idBytes) {
			idBytes[i] = b ^ idBytes[i+4]
		}
	}
	networkPart := idBytes[0:4]
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return "", err
	}

	nodePart := addressBytes[0:5]

	// Build IPv6 address from the parts
	var ipv6 [16]byte
	ipv6[0] = 0xfc
	for i := 0; i < len(networkPart); i += 1 {
		ipv6[i+1] = networkPart[i]
	}
	for i := 0; i < len(nodePart); i += 1 {
		ipv6[i+1+len(networkPart)] = nodePart[i]
	}
	ipv6[len(ipv6)-1] = 0x01

	// Format the address
	var groups [8]string
	for i := 0; i < len(ipv6); i += 2 {
		groups[i/2] = hex.EncodeToString(ipv6[i : i+2])
	}
	result := strings.Join(groups[:], ":")
	ip := net.ParseIP(result)
	return ip.String(), nil
}
