// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

func PrescanZerotierControllers(cg *client.Globals, l log.Logger) {
	for {
		controllers, err := client.GetControllers(cg.Config)
		if err != nil {
			l.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
			continue
		}

		_, err = client.ScanControllers(context.Background(), controllers, cg.Cache)
		if err != nil {
			l.Error(errors.Wrap(err, "couldn't prescan Zerotier controllers for cache"))
		}

		break
	}
}
