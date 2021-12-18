package zerotier

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
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
	eg.Go(func() error {
		s, cs, err := c.GetControllerStatuses(ctx, controller, cc)
		if err != nil {
			return err
		}
		status = s
		controllerStatus = cs
		return nil
	})
	eg.Go(func() error {
		ids, err := c.GetNetworkIDs(ctx, controller, cc)
		if err != nil {
			return err
		}
		networkIDs = ids
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return status, controllerStatus, networkIDs, nil
}
