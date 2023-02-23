package client

import (
	"context"

	"golang.org/x/sync/errgroup"

	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

func CompareSubnamesAndAddresses(
	firstSubnames []string, firstAddress string, secondSubnames []string, secondAddress string,
) bool {
	firstNamed := len(firstSubnames) > 0
	secondNamed := len(secondSubnames) > 0
	if firstNamed && secondNamed {
		// fmt.Println("Comparing subnames", firstSubnames[0], secondSubnames[0])
		return desecc.CompareSubnames(firstSubnames[0], secondSubnames[0])
	}
	if firstNamed {
		return true
	}
	if secondNamed {
		return false
	}
	return firstAddress < secondAddress
}

func GetNetworks(
	ctx context.Context, networkIDs map[string]string, c *ztc.Client, cc *ztcontrollers.Client,
) (map[string]*zerotier.ControllerNetwork, map[string]*ztcontrollers.Controller, error) {
	type IDAssociation struct {
		Subname string
		ID      string
	}
	idAssociations := make([]IDAssociation, 0, len(networkIDs))
	for subname, id := range networkIDs {
		idAssociations = append(idAssociations, IDAssociation{
			Subname: subname,
			ID:      id,
		})
	}

	eg, ctx := errgroup.WithContext(ctx)
	controllers := make([]*ztcontrollers.Controller, len(networkIDs))
	networks := make([]*zerotier.ControllerNetwork, len(networkIDs))
	for i, association := range idAssociations {
		eg.Go(func(i int, id string) func() error {
			return func() error {
				address := ztc.GetControllerAddress(id)
				controller, err := cc.FindControllerByAddress(ctx, address)
				if err != nil {
					// Tolerate unknown controllers by acting as if they (and their networks) don't exist
					return nil
				}
				controllers[i] = controller

				network, err := c.GetNetwork(ctx, *controller, id)
				if err != nil {
					return err
				}
				networks[i] = network
				return nil
			}
		}(i, association.ID))
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	// Transform the results into a more usable shape
	keyedNetworks := make(map[string]*zerotier.ControllerNetwork, len(networkIDs))
	for i, association := range idAssociations {
		keyedNetworks[association.Subname] = networks[i]
	}
	keyedControllers := make(map[string]*ztcontrollers.Controller, len(networkIDs))
	for i, association := range idAssociations {
		keyedControllers[association.Subname] = controllers[i]
	}
	return keyedNetworks, keyedControllers, nil
}
