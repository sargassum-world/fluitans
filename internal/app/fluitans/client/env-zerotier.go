package client

import (
	"os"
	"time"

	"github.com/sargassum-eco/fluitans/internal/env"
)

func GetEnvVarController() (*Controller, error) {
	url, err := env.GetURLOrigin("FLUITANS_ZT_CONTROLLER_SERVER", "", "http")
	if err != nil {
		return nil, err
	}

	var defaultNetworkCost float32 = 1.0
	networkCostWeight, err := env.GetFloat32(
		"FLUITANS_ZT_CONTROLLER_NETWORKCOST", defaultNetworkCost,
	)
	if err != nil {
		return nil, err
	}

	authtoken := os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN")
	name := env.GetString("FLUITANS_ZT_CONTROLLER_NAME", url.Host)
	desc := env.GetString(
		"FLUITANS_ZT_CONTROLLER_DESC",
		"The default ZeroTier network controller specified in the environment variables.",
	)
	if len(url.String()) == 0 || len(authtoken) == 0 {
		return nil, nil
	}

	return &Controller{
		Server:            url.String(),
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}

func GetEnvVarZerotierNetworkTTL() (int64, error) {
	var defaultTTLHours int64 = 24 // hours
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	ttl, err := env.GetInt64("FLUITANS_ZEROTIER_DNS_NETWORKTTL", defaultTTL)
	if err != nil {
		return defaultTTL, err
	}

	return ttl, nil
}

func GetEnvVarZerotierDeviceTTL() (int64, error) {
	var defaultTTLHours int64 = 1 // hour
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	ttl, err := env.GetInt64("FLUITANS_ZEROTIER_DNS_DEVICETTL", defaultTTL)
	if err != nil {
		return defaultTTL, err
	}

	return ttl, nil
}
