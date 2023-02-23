// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

func PrescanZerotierControllers(ctx context.Context, c *ztcontrollers.Client) error {
	const retryInterval = 5 * time.Second
	return handling.RepeatImmediate(ctx, retryInterval, func() (done bool, err error) {
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
	ctx context.Context, c *ztc.Client, cc *ztcontrollers.Client,
) error {
	const retryInterval = 5 * time.Second
	return handling.RepeatImmediate(ctx, retryInterval, func() (done bool, err error) {
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

func flattenRRsets(allRRsets [][]desec.RRset) []desec.RRset {
	flattened := make([]desec.RRset, 0, len(allRRsets))
	for _, rrsets := range allRRsets {
		flattened = append(flattened, rrsets...)
	}
	return flattened
}

func PlanNetworkDNSUpdates(
	ctx context.Context, controller ztcontrollers.Controller, network zerotier.ControllerNetwork,
	subnameRRsets map[string][]desec.RRset, c *ztc.Client, dc *desecc.Client,
) (rrsets []desec.RRset, err error) {
	domainName := dc.Config.DomainName
	if !client.NetworkNamedByDNS(*network.Id, *network.Name, domainName, subnameRRsets) {
		return nil, nil
	}

	memberAddresses, err := c.GetNetworkMemberAddresses(ctx, controller, *network.Id)
	if err != nil {
		return nil, err
	}
	members, err := client.GetMemberRecords(
		ctx, domainName, controller, network, memberAddresses, subnameRRsets, c,
	)
	if err != nil {
		return nil, err
	}

	upsertions := make([]desec.RRset, 0, len(members))
	for _, memberAddress := range memberAddresses {
		member := members[memberAddress]
		if len(member.DNSUpdates) == 0 {
			continue
		}
		upsertions = append(upsertions, member.ExpectedRRsets...)
	}
	return upsertions, nil
}

func PlanControllerDNSUpdates(
	ctx context.Context, controller ztcontrollers.Controller,
	networks map[string]zerotier.ControllerNetwork, subnameRRsets map[string][]desec.RRset,
	c *ztc.Client, dc *desecc.Client,
) (rrsets []desec.RRset, err error) {
	networkIDs := make([]string, 0, len(networks))
	for networkID := range networks {
		networkIDs = append(networkIDs, networkID)
	}

	eg, egctx := errgroup.WithContext(ctx)
	networkUpsertions := make([][]desec.RRset, len(networks))
	for i, networkID := range networkIDs {
		eg.Go(func(i int, networkID string) func() error {
			return func() (err error) {
				networkUpsertions[i], err = PlanNetworkDNSUpdates(
					egctx, controller, networks[networkID], subnameRRsets, c, dc,
				)
				return err
			}
		}(i, networkID))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return flattenRRsets(networkUpsertions), nil
}

func UpdateZeroTierDNSRecords(
	ctx context.Context, c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) error {
	const runInterval = 10 * time.Second
	return handling.RepeatImmediate(ctx, runInterval, func() (done bool, err error) {
		controllers, err := cc.GetControllers()
		if err != nil {
			return false, err
		}
		networkIDs, err := c.GetAllNetworkIDs(ctx, controllers, cc)
		if err != nil {
			return false, err
		}

		eg, egctx := errgroup.WithContext(ctx)
		var networks []map[string]zerotier.ControllerNetwork
		var subnameRRsets map[string][]desec.RRset
		eg.Go(func() (err error) {
			networks, err = c.GetAllNetworks(egctx, controllers, networkIDs)
			return err
		})
		eg.Go(func() (err error) {
			subnameRRsets, err = dc.GetRRsets(egctx)
			return err
		})
		if err := eg.Wait(); err != nil {
			return false, err
		}

		eg, egctx = errgroup.WithContext(ctx)
		controllerUpsertions := make([][]desec.RRset, len(controllers))
		for i, controller := range controllers {
			eg.Go(func(i int, controller ztcontrollers.Controller) func() error {
				return func() (err error) {
					controllerUpsertions[i], err = PlanControllerDNSUpdates(
						egctx, controller, networks[i], subnameRRsets, c, dc,
					)
					return err
				}
			}(i, controller))
		}
		if err := eg.Wait(); err != nil {
			return false, err
		}
		rrsets := flattenRRsets(controllerUpsertions)
		if len(rrsets) == 0 {
			return false, nil
		}

		// Apply changes
		if _, err := dc.UpsertRRsets(ctx, rrsets...); err != nil {
			return false, errors.Wrapf(err, "couldn't upsert AAAA and/or A records")
		}
		return false, nil
	})
}
