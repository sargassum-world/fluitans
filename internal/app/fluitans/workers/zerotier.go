// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/handling"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
)

func PrescanZerotierControllers(ctx context.Context, c *ztcontrollers.Client) error {
	const retryInterval = 5 * time.Second
	return handling.Repeat(ctx, retryInterval, func() (done bool, err error) {
		controllers, err := c.GetControllers()
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			return false, nil
		}

		_, err = c.ScanControllers(ctx, controllers)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prescan Zerotier controllers for cache"))
			return false, nil
		}

		return true, nil
	})
}

func PrefetchZerotierNetworks(
	ctx context.Context, c *zerotier.Client, cc *ztcontrollers.Client,
) error {
	const retryInterval = 5 * time.Second
	return handling.Repeat(ctx, retryInterval, func() (done bool, err error) {
		controllers, err := cc.GetControllers()
		if err != nil {
			cc.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			return false, nil
		}

		allNetworkIDs, err := c.GetAllNetworkIDs(ctx, controllers, cc)
		if err != nil {
			c.Logger.Error(errors.Wrap(
				err, "couldn't get the list of all Zerotier network IDs for cache",
			))
			return false, nil
		}

		_, err = c.GetAllNetworks(ctx, controllers, allNetworkIDs)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch all Zerotier networks for cache"))
			return false, nil
		}

		return true, nil
	})
}
