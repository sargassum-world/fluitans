package zerotier

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/env"
)

const envPrefix = "ZEROTIER_"

type Config struct {
	DNS ZTDNSSettings
}

func GetConfig() (c Config, err error) {
	c.DNS, err = GetDNSSettings()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Zerotier DNS settings")
	}
	return c, nil
}

func getNetworkTTL() (int64, error) {
	const defaultTTL = 60 * 60 * 24 // default: 24 hours
	return env.GetInt64(envPrefix+"DNS_NETWORKTTL", defaultTTL)
}

func getDeviceTTL() (int64, error) {
	const defaultTTL = 60 * 60 // default: 1 hour
	return env.GetInt64(envPrefix+"DNS_DEVICETTL", defaultTTL)
}

func GetDNSSettings() (s ZTDNSSettings, err error) {
	s.NetworkTTL, err = getNetworkTTL()
	if err != nil {
		return ZTDNSSettings{}, errors.Wrap(err, "couldn't make network record TTL config")
	}

	s.DeviceTTL, err = getDeviceTTL()
	if err != nil {
		return ZTDNSSettings{}, errors.Wrap(err, "couldn't make device record TTL config")
	}

	return s, nil
}
