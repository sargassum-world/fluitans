package conf

import (
	"os"
)

const dnsEnvPrefix = "DNS_"  // note: this overlaps with the prefix for the desec client

func getDomainName() string {
	return os.Getenv(dnsEnvPrefix + "DOMAIN_NAME")
}
