package zerotier

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

func (c *Client) GetControllerStatuses(
	ctx context.Context, controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) (*zerotier.Status, *zerotier.ControllerStatus, error) {
	client, err := controller.NewClient()
	if err != nil {
		return nil, nil, err
	}

	eg, ctx := errgroup.WithContext(ctx)
	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	eg.Go(func() error {
		res, cerr := client.GetStatusWithResponse(ctx)
		if cerr != nil {
			return err
		}
		status = res.JSON200
		return cc.Cache.SetControllerByAddress(*status.Address, controller)
	})
	eg.Go(func() error {
		res, cerr := client.GetControllerStatusWithResponse(ctx)
		if cerr != nil {
			return err
		}
		controllerStatus = res.JSON200
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	return status, controllerStatus, nil
}

func (c *Client) GetControllerInfo(
	ctx context.Context, controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) (*zerotier.Status, *zerotier.ControllerStatus, []string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	var networkIDs []string
	eg.Go(func() (err error) {
		status, controllerStatus, err = c.GetControllerStatuses(ctx, controller, cc)
		return
	})
	eg.Go(func() (err error) {
		networkIDs, err = c.GetNetworkIDs(ctx, controller, cc)
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return status, controllerStatus, networkIDs, nil
}
