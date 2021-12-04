// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
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
