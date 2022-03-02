package client

import (
	"fmt"
	"regexp"
	"strings"
)

func MakeNetworkIDRecord(networkID string) string {
	return fmt.Sprintf("\"zerotier-net-id=%s\"", networkID)
}

var networkIDRecordParser regexp.Regexp = *regexp.MustCompile(
	`"zerotier-net-id=([0-9a-fA-F]{16})"`,
)

func ParseNetworkIDRecord(txtRecord string) (string, bool) {
	groups := networkIDRecordParser.FindStringSubmatch(txtRecord)
	if groups == nil {
		return "", false
	}

	return strings.ToLower(groups[1]), true
}

func GetNetworkID(txtRecords []string) (string, bool) {
	for _, record := range txtRecords {
		networkID, hasID := ParseNetworkIDRecord(record)
		if hasID {
			return networkID, true
		}
	}
	return "", false
}
