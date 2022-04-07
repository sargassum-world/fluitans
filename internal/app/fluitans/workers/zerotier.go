// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
)

func PrescanZerotierControllers(ctx context.Context, c *ztcontrollers.Client) error {
	const retryInterval = 5000

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		controllers, err := c.GetControllers()
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		_, err = c.ScanControllers(ctx, controllers)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prescan Zerotier controllers for cache"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		return nil
	}
}

func PrefetchZerotierNetworks(
	ctx context.Context, c *zerotier.Client, cc *ztcontrollers.Client,
) error {
	const retryInterval = 5000

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		controllers, err := cc.GetControllers()
		if err != nil {
			cc.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		allNetworkIDs, err := c.GetAllNetworkIDs(ctx, controllers, cc)
		if err != nil {
			c.Logger.Error(errors.Wrap(
				err, "couldn't get the list of all Zerotier network IDs for cache",
			))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		_, err = c.GetAllNetworks(ctx, controllers, allNetworkIDs)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch all Zerotier networks for cache"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		return nil
	}
}
