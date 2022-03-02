package desec

import (
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/models"
	"github.com/sargassum-eco/fluitans/pkg/godest/env"
)

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
		err = errors.Wrap(err, "couldn't make DNS server config")
		return
	}

	c.APISettings, err = GetAPISettings()
	if err != nil {
		err = errors.Wrap(err, "couldn't make deSEC API settings")
		return
	}

	c.RecordTypes = getRecordTypes()
	return
}

func getDNSServer() (s models.DNSServer, err error) {
	url, err := env.GetURLOrigin("FLUITANS_DNS_SERVER", "", "https")
	if err != nil {
		err = errors.Wrap(err, "couldn't make server url config")
		return
	}
	s.Server = url.String()
	if len(s.Server) == 0 {
		s = models.DNSServer{}
		return
	}
	s.API = strings.ToLower(env.GetString("FLUITANS_DNS_API", "desec"))
	if len(s.API) == 0 {
		s = models.DNSServer{}
		return
	}

	s.Authtoken = os.Getenv("FLUITANS_DNS_AUTHTOKEN")
	if len(s.Authtoken) == 0 {
		s = models.DNSServer{}
		return
	}

	s.Name = env.GetString("FLUITANS_DNS_NAME", url.Host)
	s.Description = env.GetString(
		"FLUITANS_DNS_DESC",
		"The default deSEC DNS server account specified in the environment variables.",
	)

	const defaultNetworkCost = 2.0
	s.NetworkCostWeight, err = env.GetFloat32("FLUITANS_DNS_NETWORKCOST", defaultNetworkCost)
	if err != nil {
		err = errors.Wrap(err, "couldn't make network cost config")
		return
	}
	return
}

func getReadCacheTTL() (time.Duration, error) {
	// The TTL of cached read results, in units of seconds; negative numbers represent an infinite
	// TTL. All reads are cached, and cache entries below TTL will be used instead of issuing extra
	// API read requests. The cache will be consistent with the API at an infinite TTL (the default
	// TTL) if we promise to only modify DNS records through Fluitans, and not independently through
	// the deSEC server.
	rawTTL, err := env.GetFloat32("FLUITANS_DESEC_READ_CACHE_TTL", -1)
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
		err = errors.Wrap(err, "couldn't make readCacheTTL config")
		return
	}

	// The write limiter fill ratio above which RRset writes, rather than being executed immediately,
	// will first be batched into groups based on the nearest rate limit
	const defaultWriteSoftQuota = 0.34
	s.WriteSoftQuota, err = env.GetFloat32("FLUITANS_DESEC_WRITE_SOFT_QUOTA", defaultWriteSoftQuota)
	if err != nil {
		err = errors.Wrap(err, "couldn't make writeSoftQuota config")
		return
	}
	if s.WriteSoftQuota < 0 {
		s.WriteSoftQuota = 0
	}
	if s.WriteSoftQuota > 1 {
		s.WriteSoftQuota = 1
	}

	return
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
