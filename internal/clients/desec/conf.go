package desec

import (
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/env"

	"github.com/sargassum-world/fluitans/internal/models"
)

const envPrefix = "DNS_"

type Config struct {
	DomainName  string
	DNSServer   models.DNSServer
	APISettings DesecAPISettings
	RecordTypes []string
}

func GetConfig(domainName string) (c Config, err error) {
	c.DomainName = domainName
	c.DNSServer, err = getDNSServer()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make DNS server config")
	}

	c.APISettings, err = GetAPISettings()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make deSEC API settings")
	}

	c.RecordTypes = getRecordTypes()
	return c, nil
}

func getDNSServer() (s models.DNSServer, err error) {
	url, err := env.GetURLOrigin(envPrefix+"SERVER", "", "https")
	if err != nil {
		return models.DNSServer{}, errors.Wrap(err, "couldn't make server url config")
	}
	s.Server = url.String()
	if len(s.Server) == 0 {
		return models.DNSServer{}, nil
	}
	s.API = strings.ToLower(env.GetString(envPrefix+"API", "desec"))
	if len(s.API) == 0 {
		return models.DNSServer{}, nil
	}

	s.Authtoken = os.Getenv(envPrefix + "AUTHTOKEN")
	if len(s.Authtoken) == 0 {
		return models.DNSServer{}, nil
	}

	s.Name = env.GetString(envPrefix+"NAME", url.Host)
	s.Description = env.GetString(
		envPrefix+"DESC",
		"The default deSEC DNS server account specified in the environment variables.",
	)

	const defaultNetworkCost = 2.0
	s.NetworkCostWeight, err = env.GetFloat32(envPrefix+"NETWORKCOST", defaultNetworkCost)
	if err != nil {
		return models.DNSServer{}, errors.Wrap(err, "couldn't make network cost config")
	}
	return s, nil
}

func getReadCacheTTL() (time.Duration, error) {
	// The TTL of cached read results, in units of seconds; negative numbers represent an infinite
	// TTL. All reads are cached, and cache entries below TTL will be used instead of issuing extra
	// API read requests. The cache will be consistent with the API at an infinite TTL (the default
	// TTL) if we promise to only modify DNS records through Fluitans, and not independently through
	// the deSEC server.
	const defaultTTL = 60 * 10 // default: 10 minutes
	rawTTL, err := env.GetFloat32(envPrefix+"READ_CACHE_TTL", defaultTTL)
	var ttl time.Duration = -1
	if rawTTL >= 0 {
		durationReadCacheTTL := time.Duration(rawTTL) * time.Second
		ttl = durationReadCacheTTL
	}
	if err != nil {
		return -1, err
	}

	return ttl, nil
}

func GetAPISettings() (s DesecAPISettings, err error) {
	s.ReadCacheTTL, err = getReadCacheTTL()
	if err != nil {
		return DesecAPISettings{}, errors.Wrap(err, "couldn't make readCacheTTL config")
	}

	// The write limiter fill ratio above which RRset writes, rather than being executed immediately,
	// will first be batched into groups based on the nearest rate limit
	const defaultWriteSoftQuota = 0.34
	s.WriteSoftQuota, err = env.GetFloat32(envPrefix+"WRITE_SOFT_QUOTA", defaultWriteSoftQuota)
	if err != nil {
		return DesecAPISettings{}, errors.Wrap(err, "couldn't make writeSoftQuota config")
	}
	if s.WriteSoftQuota < 0 {
		s.WriteSoftQuota = 0
	}
	if s.WriteSoftQuota > 1 {
		s.WriteSoftQuota = 1
	}
	return s, nil
}

func getRecordTypes() []string {
	return []string{
		"A",
		"AAAA",
		// "CAA",
		// "CERT",
		"CNAME",
		"DNAME",
		"LOC",
		"NS",
		"PTR",
		"RP",
		"SRV",
		// "SSHFP",
		// "TLSA",
		"TXT",
		"URI",
	}
}
