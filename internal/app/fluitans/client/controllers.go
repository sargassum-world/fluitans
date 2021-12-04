package client

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
)

func GetControllers(config conf.Config) ([]models.Controller, error) {
	// TODO: look up the controllers from a database, if one is specified!
	controllers := make([]models.Controller, 0)
	envController := config.Controller

	if envController != nil {
		controllers = append(controllers, *envController)
	}
	return controllers, nil
}

func FindController(
	name string, config conf.Config,
) (*models.Controller, bool, error) {
	found := false
	controllers, err := GetControllers(config)
	if err != nil {
		return nil, false, err
	}

	for _, v := range controllers {
		if v.Name == name {
			return &v, true, nil
		}
	}
	return nil, found, nil
}
