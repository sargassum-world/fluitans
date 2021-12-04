// Package conf supports environment variable-based application configuration
package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type Config struct {
	Cache      ristretto.Config
	DomainName string
}

func GetConfig() (*Config, error) {
	cacheConfig, err := getCacheConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make cache config")
	}

	return &Config{
		Cache:      *cacheConfig,
		DomainName: getDomainName(),
	}, nil
}
