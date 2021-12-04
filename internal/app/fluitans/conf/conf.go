// Package conf supports environment variable-based application configuration
package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
)

type Config struct {
	Cache       ristretto.Config
	DesecAPI    models.DesecAPISettings
	DomainName  string
	DNSServer   models.DNSServer
	Controller  *models.Controller
	ZerotierDNS models.ZerotierDNSSettings
}

func GetConfig() (*Config, error) {
	cacheConfig, err := getCacheConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make cache config")
	}

	desecAPISettings, err := getDesecAPISettings()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make deSEC API settings")
	}

	domainName := getDomainName()

	dnsServer, err := getDNSServer()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make DNS server config")
	}

	controller, err := getController()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Zerotier controller config")
	}

	zerotierDNSSettings, err := getZerotierDNSSettings()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Zerotier DNS settings")
	}

	return &Config{
		Cache:       *cacheConfig,
		DesecAPI:    *desecAPISettings,
		DomainName:  domainName,
		DNSServer:   *dnsServer,
		Controller:  controller,
		ZerotierDNS: *zerotierDNSSettings,
	}, nil
}
