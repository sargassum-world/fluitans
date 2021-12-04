package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/internal/log"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// Controller

func checkCachedController(
	ctx context.Context, address string, cache *Cache, l log.Logger,
) (*models.Controller, error) {
	controller, cacheHit, err := cache.GetControllerByAddress(address)
	if err != nil {
		return nil, err
	}

	if !cacheHit {
		return nil, nil
	}

	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		// evict the cached controller if its authtoken is stale or it no longer responds
		cache.UnsetControllerByAddress(address)
		return nil, err
	}

	if address != *res.JSON200.Address { // cached controller's address is stale
		l.Warnf(
			"controller %s's address has changed from %s to %s",
			controller.Server, address, *res.JSON200.Address,
		)
		err = cache.SetControllerByAddress(*res.JSON200.Address, *controller)
		if err != nil {
			return nil, err
		}

		cache.UnsetControllerByAddress(address)
		return nil, nil
	}

	return controller, nil
}

func ScanControllers(
	ctx context.Context, controllers []models.Controller, cache *Cache,
) ([]string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	addresses := make([]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(
					controller.Server, controller.Authtoken,
				)
				if cerr != nil {
					return nil
				}

				res, err := client.GetStatusWithResponse(ctx)
				if err != nil {
					return err
				}

				addresses[i] = *res.JSON200.Address
				return nil
			}
		}(i))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for i, v := range controllers {
		err := cache.SetControllerByAddress(addresses[i], v)
		if err != nil {
			return nil, err
		}
	}

	return addresses, nil
}

func FindControllerByAddress(
	ctx context.Context, address string,
	config conf.Config, cache *Cache, l log.Logger,
) (*models.Controller, error) {
	controller, err := checkCachedController(ctx, address, cache, l)
	if err != nil {
		// Log the error and proceed to manually query all controllers
		l.Error(err, errors.Wrap(err, fmt.Sprintf(
			"couldn't handle the cache entry for the controller with address %s", address,
		)))
	} else if controller != nil {
		return controller, nil
	}

	// Query the list of all known controllers
	controllers, err := GetControllers(config)
	if err != nil {
		return nil, err
	}

	l.Warnf(
		"rescanning controllers due to a stale/missing controller for %s in cache",
		address,
	)
	addresses, err := ScanControllers(ctx, controllers, cache)
	if err != nil {
		return nil, err
	}

	for i, v := range controllers {
		if addresses[i] == address {
			return &v, nil
		}
	}

	return nil, echo.NewHTTPError(
		http.StatusNotFound,
		fmt.Sprintf("Controller not found with address %s", address),
	)
}

func GetController(
	ctx context.Context, controller models.Controller, cache *Cache,
) (*zerotier.Status, *zerotier.ControllerStatus, []string, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if cerr != nil {
		return nil, nil, nil, cerr
	}

	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	var networks []string
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		res, err := client.GetStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		status = res.JSON200
		if err := cache.SetControllerByAddress(*status.Address, controller); err != nil {
			return err
		}

		return nil
	})
	eg.Go(func() error {
		res, err := client.GetControllerStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		controllerStatus = res.JSON200
		return err
	})
	eg.Go(func() error {
		res, err := client.GetControllerNetworksWithResponse(ctx)
		if err != nil {
			return err
		}

		networks = *res.JSON200
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return status, controllerStatus, networks, nil
}
