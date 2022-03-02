// Package conf supports environment variable-based application configuration
package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type Config struct {
	Cache      ristretto.Config
	DomainName string
	HTTP       HTTPConfig
}

func GetConfig() (c Config, err error) {
	c.Cache, err = getCacheConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't make cache config")
		return
	}

	c.DomainName = getDomainName()

	c.HTTP, err = getHTTPConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't make http config")
		return
	}

	return
}
