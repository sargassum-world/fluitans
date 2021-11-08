// Package client contains client code for external APIs
package client

import (
	"net/url"
	"os"
)

func getEnvVarController() (*Controller, error) {
	rawURL := os.Getenv("FLUITANS_ZT_CONTROLLER_SERVER")
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if len(parsedURL.Scheme) == 0 {
		parsedURL.Scheme = "http"
	}
	parsedURL.Path = ""
	parsedURL.User = nil
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	authtoken := os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN")
	name := os.Getenv("FLUITANS_ZT_CONTROLLER_NAME")
	if len(name) == 0 {
		name = parsedURL.Host
	}
	desc := os.Getenv("FLUITANS_ZT_CONTROLLER_DESC")
	if len(desc) == 0 {
		desc = "The default ZeroTier network controller specified in the environment variables."
	}
	if len(parsedURL.String()) == 0 || len(authtoken) == 0 {
		return nil, nil
	}

	return &Controller {
		Server:      parsedURL.String(),
		Name:        name,
		Description: desc,
		Authtoken:   authtoken,
	}, nil
}

func GetDomainName() string {
	return os.Getenv("FLUITANS_DOMAIN_NAME")
}
