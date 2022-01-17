package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func addNDPAddresses(
	id string, v6AssignMode zerotier.V6AssignMode,
	members map[string]zerotier.ControllerNetworkMember,
) error {
	n6plane := v6AssignMode.N6plane
	if n6plane != nil && *n6plane {
		for _, member := range members {
			n6PlaneAddress, err := zerotier.Get6Plane(id, *member.Address)
			if err != nil {
				return err
			}

			if member.IpAssignments == nil {
				ipAssignments := []string{n6PlaneAddress}
				member.IpAssignments = &ipAssignments
			} else {
				*member.IpAssignments = append(*member.IpAssignments, n6PlaneAddress)
			}
		}
	}
	return nil
}

func checkNamedByDNS(
	ctx context.Context, networkName, networkID string, c *desec.Client,
) (bool, error) {
	domainSuffix := fmt.Sprintf(".%s", c.Config.DomainName)
	if !strings.HasSuffix(networkName, domainSuffix) {
		return false, nil
	}

	subname := strings.TrimSuffix(networkName, domainSuffix)
	txtRRset, err := c.GetRRset(ctx, subname, "TXT")
	if err != nil {
		return false, err
	}

	if txtRRset != nil {
		parsedID, txtHasNetworkID := client.GetNetworkID(txtRRset.Records)
		return txtHasNetworkID && (parsedID == networkID), nil
	}

	return false, nil
}

type NetworkData struct {
	Controller       ztcontrollers.Controller
	Network          zerotier.ControllerNetwork
	Members          map[string]zerotier.ControllerNetworkMember
	JSONPrintedRules string
	DomainName       string
	NamedByDNS       bool
}

func getNetworkData(
	ctx context.Context, address, id string,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desec.Client,
) (*NetworkData, error) {
	controller, err := cc.FindControllerByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	network, memberAddresses, err := c.GetNetworkInfo(ctx, *controller, id)
	if err != nil {
		return nil, err
	}
	rules, err := json.MarshalIndent(*network.Rules, "", "  ")
	if err != nil {
		return nil, err
	}

	members, err := c.GetNetworkMembersInfo(ctx, *controller, id, memberAddresses)
	if err != nil {
		return nil, err
	}
	if err = addNDPAddresses(id, *network.V6AssignMode, members); err != nil {
		return nil, err
	}

	namedByDNS, err := checkNamedByDNS(ctx, *network.Name, *network.Id, dc)
	if err != nil {
		return nil, err
	}

	// TODO: also retrieve and get other records to render on the page
	return &NetworkData{
		Controller:       *controller,
		Network:          *network,
		Members:          members,
		JSONPrintedRules: string(rules),
		DomainName:       dc.Config.DomainName,
		NamedByDNS:       namedByDNS,
	}, nil
}

func getNetwork(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/network.page.tmpl"
	err := te.RequireSegments("networks.getNetwork", t)
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			id := c.Param("id")
			address := ztc.GetControllerAddress(id)

			// Run queries
			networkData, err := getNetworkData(
				ctx, address, id, app.Clients.Zerotier, app.Clients.ZTControllers, app.Clients.Desec,
			)
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, *networkData, te, g)
		}, nil
	}
}

func nameNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string, name string,
	c *ztc.Client, dc *desec.Client,
) error {
	if len(name) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot remove name from network")
	}

	// Check to see if the network was already named by DNS
	fqdn := fmt.Sprintf("%s.%s", name, dc.Config.DomainName)
	txtRRset, err := dc.GetRRset(ctx, name, "TXT")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't check cache for DNS TXT RRset at %s for network %s", fqdn, id,
		))
	}
	if txtRRset != nil {
		if _, hasID := client.GetNetworkID(txtRRset.Records); hasID {
			return echo.NewHTTPError(http.StatusBadRequest, "name is already used by another network")
		}
	}

	ttl := c.Config.DNS.NetworkTTL
	if _, err := dc.CreateRRset(
		ctx, name, "TXT", ttl, []string{client.MakeNetworkIDRecord(id)},
	); err != nil {
		// TODO: if a TXT RRset already exists but doesn't have the ID, just append a
		// zerotier-net-id=... record (but we should have a global lock on a get-and-patch to avoid
		// data races)
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't create a DNS TXT RRset at %s for network %s", fqdn, id,
		))
	}
	return c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
}

func setNetworkRules(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, jsonRules string, c *ztc.Client,
) error {
	rules := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return err
	}
	err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	if err != nil {
		return err
	}
	return nil
}

func postNetwork(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		zc := app.Clients.Zerotier
		cc := app.Clients.ZTControllers
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			id := c.Param("id")
			address := ztc.GetControllerAddress(id)
			method := c.FormValue("method")

			// Run queries
			controller, err := cc.FindControllerByAddress(ctx, address)
			if err != nil {
				return err
			}

			switch method {
			default:
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"invalid POST method %s", method,
				))
			case "SETNAME":
				if err = nameNetwork(
					ctx, *controller, id, c.FormValue("name"), app.Clients.Zerotier, app.Clients.Desec,
				); err != nil {
					return err
				}
			case "SETRULES":
				if err = setNetworkRules(ctx, *controller, id, c.FormValue("rules"), zc); err != nil {
					return err
				}
				return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id))
			case "DELETE":
				if err = zc.DeleteNetwork(ctx, *controller, id, app.Clients.ZTControllers); err != nil {
					// TODO: add a tombstone to the TXT RRset?
					return err
				}
				return c.Redirect(http.StatusSeeOther, "/networks")
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
		}, nil
	}
}
