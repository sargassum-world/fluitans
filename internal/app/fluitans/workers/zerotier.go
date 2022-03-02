// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
)

func PrescanZerotierControllers(c *ztcontrollers.Client) {
	for {
		controllers, err := c.GetControllers()
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			continue
		}

		_, err = c.ScanControllers(context.Background(), controllers)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prescan Zerotier controllers for cache"))
		}

		break
	}
}

func PrefetchZerotierNetworks(c *zerotier.Client, cc *ztcontrollers.Client) {
	for {
		controllers, err := cc.GetControllers()
		if err != nil {
			cc.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			continue
		}

		allNetworkIDs, err := c.GetAllNetworkIDs(context.Background(), controllers, cc)
		if err != nil {
			c.Logger.Error(errors.Wrap(
				err, "couldn't get the list of all Zerotier network IDs for cache",
			))
		}

		_, err = c.GetAllNetworks(context.Background(), controllers, allNetworkIDs)
		if err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch all Zerotier networks for cache"))
		}

		break
	}
}
