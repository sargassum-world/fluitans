package ztcontrollers

import (
	"os"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/env"
)

type Config struct {
	Controller *Controller
}

func GetConfig() (*Config, error) {
	controller, err := GetController()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Zerotier controller config")
	}

	return &Config{
		Controller: controller,
	}, nil
}

func GetController() (*Controller, error) {
	url, err := env.GetURLOrigin("FLUITANS_ZT_CONTROLLER_SERVER", "", "http")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make server url config")
	}

	var defaultNetworkCost float32 = 1.0
	networkCostWeight, err := env.GetFloat32("FLUITANS_ZT_CONTROLLER_NETWORKCOST", defaultNetworkCost)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make network cost config")
	}

	authtoken := os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN")
	name := env.GetString("FLUITANS_ZT_CONTROLLER_NAME", url.Host)
	desc := env.GetString(
		"FLUITANS_ZT_CONTROLLER_DESC",
		"The default ZeroTier network controller specified in the environment variables.",
	)
	if len(url.String()) == 0 || len(authtoken) == 0 {
		return nil, nil
	}

	return &Controller{
		Server:            url.String(),
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}
