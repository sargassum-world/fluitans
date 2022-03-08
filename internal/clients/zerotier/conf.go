package zerotier

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "ZEROTIER_"

type Config struct {
	DNS ZTDNSSettings
}

func GetConfig() (c Config, err error) {
	c.DNS, err = GetDNSSettings()
	if err != nil {
		err = errors.Wrap(err, "couldn't make Zerotier DNS settings")
		return
	}
	return
}

func getNetworkTTL() (int64, error) {
	const defaultTTL = 24 * 60 * 60 // default: 24 hours
	return env.GetInt64(envPrefix+"DNS_NETWORKTTL", defaultTTL)
}

func getDeviceTTL() (int64, error) {
	const defaultTTL = 1 * 60 * 60 // default: 1 hour
	return env.GetInt64(envPrefix+"DNS_DEVICETTL", defaultTTL)
}

func GetDNSSettings() (s ZTDNSSettings, err error) {
	s.NetworkTTL, err = getNetworkTTL()
	if err != nil {
		err = errors.Wrap(err, "couldn't make network record TTL config")
		return
	}

	s.DeviceTTL, err = getDeviceTTL()
	if err != nil {
		err = errors.Wrap(err, "couldn't make device record TTL config")
		return
	}

	return
}
