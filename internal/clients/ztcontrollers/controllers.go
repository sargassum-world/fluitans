package ztcontrollers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// All Controllers

func (c *Client) GetControllers() ([]Controller, error) {
	// TODO: look up the controllers from a database, if one is specified!
	controllers := make([]Controller, 0)
	envController := c.Config.Controller

	if envController != (Controller{}) {
		controllers = append(controllers, envController)
	}
	return controllers, nil
}

func (c *Client) ScanControllers(ctx context.Context, controllers []Controller) ([]string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	addresses := make([]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int) func() error {
			return func() error {
				client, cerr := controller.NewClient()
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
		if err := c.Cache.SetControllerByAddress(addresses[i], v); err != nil {
			return nil, err
		}
		if err := c.Cache.SetAddressByServer(v.Server, addresses[i], v.NetworkCostWeight); err != nil {
			return nil, err
		}
	}

	return addresses, nil
}

// Individual Controller

func (c *Client) FindController(name string) (*Controller, error) {
	controllers, err := c.GetControllers()
	if err != nil {
		return nil, err
	}

	for _, v := range controllers {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, nil
}

func (c *Client) checkCachedController(ctx context.Context, address string) (*Controller, error) {
	controller, cacheHit, err := c.Cache.GetControllerByAddress(address)
	if err != nil {
		return nil, err
	}

	if !cacheHit {
		return nil, nil
	}

	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		// evict the cached controller if its authtoken is stale or it no longer responds
		c.Cache.UnsetControllerByAddress(address)
		return nil, err
	}

	if address != *res.JSON200.Address { // cached controller's address is stale
		c.Logger.Warnf(
			"zerotier controller %s's address has changed from %s to %s",
			controller.Server, address, *res.JSON200.Address,
		)
		err = c.Cache.SetControllerByAddress(*res.JSON200.Address, *controller)
		if err != nil {
			return nil, err
		}

		c.Cache.UnsetControllerByAddress(address)
		c.Cache.UnsetAddressByServer(controller.Server)
		return nil, nil
	}

	return controller, nil
}

func (c *Client) FindControllerByAddress(ctx context.Context, address string) (*Controller, error) {
	controller, err := c.checkCachedController(ctx, address)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Log the error and proceed to manually query all controllers
		c.Logger.Error(err, errors.Wrapf(
			err, "couldn't handle the cache entry for the zerotier controller with address %s", address,
		))
	} else if controller != nil {
		return controller, nil
	}

	// Query the list of all known controllers
	controllers, err := c.GetControllers()
	if err != nil {
		return nil, err
	}

	c.Logger.Warnf(
		"rescanning zerotier controllers due to a stale/missing controller for %s in cache", address,
	)
	addresses, err := c.ScanControllers(ctx, controllers)
	if err != nil {
		return nil, err
	}

	for i, v := range controllers {
		if addresses[i] == address {
			return &v, nil
		}
	}

	return nil, echo.NewHTTPError(
		http.StatusNotFound, fmt.Sprintf("zerotier controller not found with address %s", address),
	)
}

func (c *Client) getAddressFromCache(controller Controller) (string, bool) {
	address, cacheHit, err := c.Cache.GetAddressByServer(controller.Server)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Log the error but return as a cache miss so we can manually query the controller
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the Zerotier address for %s", controller.Server,
		))
		return "", false // treat an unparseable cache entry like a cache miss
	}

	return address, cacheHit
}

func (c *Client) getAddressFromZerotier(
	ctx context.Context, controller Controller,
) (string, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return "", cerr
	}

	sRes, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return "", err
	}
	status := *sRes.JSON200
	if err := c.Cache.SetAddressByServer(
		controller.Server, *status.Address, controller.NetworkCostWeight,
	); err != nil {
		return "", err
	}

	return *status.Address, nil
}

func (c *Client) GetAddress(ctx context.Context, controller Controller) (string, error) {
	if address, cacheHit := c.getAddressFromCache(controller); cacheHit {
		return address, nil // empty address indicates nonexistent address
	}
	return c.getAddressFromZerotier(ctx, controller)
}
