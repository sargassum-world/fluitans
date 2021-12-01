package client

import (
	"os"
	"strings"

	"github.com/sargassum-eco/fluitans/internal/env"
)

func GetEnvVarDomainName() string {
	return os.Getenv("FLUITANS_DOMAIN_NAME")
}

func GetEnvVarDNSServer() (*DNSServer, error) {
	url, err := env.GetURLOrigin("FLUITANS_DNS_SERVER", "", "https")
	if err != nil {
		return nil, err
	}

	var defaultNetworkCost float32 = 2.0
	networkCostWeight, err := env.GetFloat32(
		"FLUITANS_DNS_NETWORKCOST", defaultNetworkCost,
	)
	if err != nil {
		return nil, err
	}

	api := strings.ToLower(env.GetString("FLUITANS_DNS_API", "desec"))
	authtoken := os.Getenv("FLUITANS_DNS_AUTHTOKEN")
	name := env.GetString("FLUITANS_DNS_NAME", url.Host)
	desc := env.GetString(
		"FLUITANS_DNS_DESC",
		"The default deSEC DNS server account specified in the environment variables.",
	)
	if len(url.String()) == 0 || len(authtoken) == 0 {
		return nil, nil
	}

	return &DNSServer{
		Server:            url.String(),
		API:               api,
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}
