package ztcontrollers

import (
	"os"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "ZTCONTROLLER_"

type Config struct {
	Controller Controller
}

func GetConfig() (c Config, err error) {
	c.Controller, err = GetController()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Zerotier controller config")
	}
	return c, nil
}

func GetController() (c Controller, err error) {
	url, err := env.GetURLOrigin(envPrefix+"SERVER", "", "http")
	if err != nil {
		return Controller{}, errors.Wrap(err, "couldn't make server url config")
	}
	c.Server = url.String()
	if len(c.Server) == 0 {
		return Controller{}, nil
	}

	c.Authtoken = os.Getenv(envPrefix + "AUTHTOKEN")
	if len(c.Authtoken) == 0 {
		return Controller{}, nil
	}

	c.Name = env.GetString(envPrefix+"NAME", url.Host)
	c.Description = env.GetString(
		envPrefix+"DESC",
		"The default ZeroTier network controller specified in the environment variables.",
	)

	const defaultNetworkCost = 1.0
	c.NetworkCostWeight, err = env.GetFloat32(envPrefix+"NETWORKCOST", defaultNetworkCost)
	if err != nil {
		return Controller{}, errors.Wrap(err, "couldn't make network cost config")
	}
	return c, nil
}
