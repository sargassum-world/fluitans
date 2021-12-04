package conf

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/pkg/framework/env"
)

func getDomainName() string {
	return os.Getenv("FLUITANS_DOMAIN_NAME")
}

func getDNSServer() (*models.DNSServer, error) {
	url, err := env.GetURLOrigin("FLUITANS_DNS_SERVER", "", "https")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make server url config")
	}

	var defaultNetworkCost float32 = 2.0
	networkCostWeight, err := env.GetFloat32(
		"FLUITANS_DNS_NETWORKCOST", defaultNetworkCost,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make network cost config")
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

	return &models.DNSServer{
		Server:            url.String(),
		API:               api,
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}
