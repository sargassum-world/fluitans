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

func GetConfig(domainName string) (*Config, error) {
	dnsServer, err := getDNSServer()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make DNS server config")
	}

	apiSettings, err := GetAPISettings()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make deSEC API settings")
	}

	return &Config{
		DomainName:  domainName,
		DNSServer:   *dnsServer,
		APISettings: *apiSettings,
		RecordTypes: getRecordTypes(),
	}, nil
}

func getDNSServer() (*models.DNSServer, error) {
	url, err := env.GetURLOrigin("FLUITANS_DNS_SERVER", "", "https")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make server url config")
	}

	var defaultNetworkCost float32 = 2.0
	networkCostWeight, err := env.GetFloat32("FLUITANS_DNS_NETWORKCOST", defaultNetworkCost)
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

func GetAPISettings() (*DesecAPISettings, error) {
	readCacheTTL, err := getReadCacheTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make readCacheTTL config")
	}

	// The write limiter fill ratio above which RRset writes, rather than being executed immediately,
	// will first be batched into groups based on the nearest rate limit
	var defaultWriteSoftQuota float32 = 0.34
	writeSoftQuota, err := env.GetFloat32("FLUITANS_DESEC_WRITE_SOFT_QUOTA", defaultWriteSoftQuota)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make writeSoftQuota config")
	}

	if writeSoftQuota < 0 {
		writeSoftQuota = 0
	}
	if writeSoftQuota > 1 {
		writeSoftQuota = 1
	}

	return &DesecAPISettings{
		ReadCacheTTL:   readCacheTTL,
		WriteSoftQuota: writeSoftQuota,
	}, nil
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
