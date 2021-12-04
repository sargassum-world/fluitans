package conf

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/pkg/framework/env"
)

func getController() (*models.Controller, error) {
	url, err := env.GetURLOrigin("FLUITANS_ZT_CONTROLLER_SERVER", "", "http")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make server url config")
	}

	var defaultNetworkCost float32 = 1.0
	networkCostWeight, err := env.GetFloat32(
		"FLUITANS_ZT_CONTROLLER_NETWORKCOST", defaultNetworkCost,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make network cost config")
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

	return &models.Controller{
		Server:            url.String(),
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}

func getZerotierNetworkTTL() (int64, error) {
	var defaultTTLHours int64 = 24 // hours
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	return env.GetInt64("FLUITANS_ZEROTIER_DNS_NETWORKTTL", defaultTTL)
}

func getZerotierDeviceTTL() (int64, error) {
	var defaultTTLHours int64 = 1 // hour
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	return env.GetInt64("FLUITANS_ZEROTIER_DNS_DEVICETTL", defaultTTL)
}

func getZerotierDNSSettings() (*models.ZerotierDNSSettings, error) {
	networkTTL, err := getZerotierNetworkTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make network record TTL config")
	}

	deviceTTL, err := getZerotierDeviceTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make device record TTL config")
	}

	return &models.ZerotierDNSSettings{
		NetworkTTL: networkTTL,
		DeviceTTL:  deviceTTL,
	}, nil
}
