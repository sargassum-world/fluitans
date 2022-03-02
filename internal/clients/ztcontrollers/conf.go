package ztcontrollers

import (
	"os"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/godest/env"
)

type Config struct {
	Controller Controller
}

func GetConfig() (c Config, err error) {
	c.Controller, err = GetController()
	if err != nil {
		err = errors.Wrap(err, "couldn't make Zerotier controller config")
		return
	}

	return
}

func GetController() (c Controller, err error) {
	url, err := env.GetURLOrigin("FLUITANS_ZT_CONTROLLER_SERVER", "", "http")
	if err != nil {
		err = errors.Wrap(err, "couldn't make server url config")
		return
	}
	c.Server = url.String()
	if len(c.Server) == 0 {
		c = Controller{}
		return
	}

	c.Authtoken = os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN")
	if len(c.Authtoken) == 0 {
		c = Controller{}
		return
	}

	c.Name = env.GetString("FLUITANS_ZT_CONTROLLER_NAME", url.Host)
	c.Description = env.GetString(
		"FLUITANS_ZT_CONTROLLER_DESC",
		"The default ZeroTier network controller specified in the environment variables.",
	)

	const defaultNetworkCost = 1.0
	c.NetworkCostWeight, err = env.GetFloat32("FLUITANS_ZT_CONTROLLER_NETWORKCOST", defaultNetworkCost)
	if err != nil {
		err = errors.Wrap(err, "couldn't make network cost config")
		return
	}

	return
}
