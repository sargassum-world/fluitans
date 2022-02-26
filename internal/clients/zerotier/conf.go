package zerotier

import (
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/godest/env"
)

type Config struct {
	DNS ZTDNSSettings
}

func GetConfig() (*Config, error) {
	zerotierDNSSettings, err := GetDNSSettings()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Zerotier DNS settings")
	}

	return &Config{
		DNS: *zerotierDNSSettings,
	}, nil
}

func getNetworkTTL() (int64, error) {
	var defaultTTLHours int64 = 24 // hours
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	return env.GetInt64("FLUITANS_ZEROTIER_DNS_NETWORKTTL", defaultTTL)
}

func getDeviceTTL() (int64, error) {
	var defaultTTLHours int64 = 1 // hour
	defaultTTL := int64((time.Duration(defaultTTLHours) * time.Hour).Seconds())
	return env.GetInt64("FLUITANS_ZEROTIER_DNS_DEVICETTL", defaultTTL)
}

func GetDNSSettings() (*ZTDNSSettings, error) {
	networkTTL, err := getNetworkTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make network record TTL config")
	}

	deviceTTL, err := getDeviceTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make device record TTL config")
	}

	return &ZTDNSSettings{
		NetworkTTL: networkTTL,
		DeviceTTL:  deviceTTL,
	}, nil
}
