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
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func addNDPAddresses(
	id string,
	v6AssignMode zerotier.V6AssignMode,
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
	ctx context.Context, networkName, networkID string,
	cg *client.Globals, l echo.Logger,
) (bool, error) {
	domain := cg.DNSDomain
	domainSuffix := fmt.Sprintf(".%s", domain.DomainName)
	if !strings.HasSuffix(networkName, domainSuffix) {
		return false, nil
	}

	subname := strings.TrimSuffix(networkName, domainSuffix)
	txtRRset, err := client.GetRRset(ctx, domain, subname, "TXT", l)
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
	Controller       models.Controller
	Network          zerotier.ControllerNetwork
	Members          map[string]zerotier.ControllerNetworkMember
	JSONPrintedRules string
	DomainName       string
	NamedByDNS       bool
}

func getNetworkData(
	ctx context.Context, address string, id string,
	cg *client.Globals, l echo.Logger,
) (*NetworkData, error) {
	controller, err := client.FindControllerByAddress(
		ctx, address, cg.Config, cg.Cache, l,
	)
	if err != nil {
		return nil, err
	}

	network, memberAddresses, err := client.GetNetworkInfo(ctx, *controller, id)
	if err != nil {
		return nil, err
	}

	members, err := client.GetNetworkMembersInfo(ctx, *controller, id, memberAddresses)
	if err != nil {
		return nil, err
	}

	err = addNDPAddresses(id, *network.V6AssignMode, members)
	if err != nil {
		return nil, err
	}

	rules, err := json.MarshalIndent(*network.Rules, "", "  ")
	if err != nil {
		return nil, err
	}

	namedByDNS, err := checkNamedByDNS(ctx, *network.Name, *network.Id, cg, l)
	if err != nil {
		return nil, err
	}

	// TODO: also retrieve and get other records to render on the page
	return &NetworkData{
		Controller:       *controller,
		Network:          *network,
		Members:          members,
		JSONPrintedRules: string(rules),
		DomainName:       cg.Config.DomainName,
		NamedByDNS:       namedByDNS,
	}, nil
}

func getNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/network.page.tmpl"
	tte, err := templates.GetTemplate(te, t, "networks.getNetwork")
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()
			l := c.Logger()

			// Parse params
			id := c.Param("id")
			address := client.GetControllerAddress(id)

			// Run queries
			networkData, err := getNetworkData(ctx, address, id, app, l)
			if err != nil {
				return err
			}

			// Produce output
			noContent, err := templates.ProcessEtag(c, tte, networkData)
			if err != nil || noContent {
				return err
			}
			return c.Render(http.StatusOK, t, templates.MakeRenderData(c, g, *networkData))
		}, nil
	}
}

func nameNetwork(
	ctx context.Context, controller models.Controller, id string, name string,
	cg *client.Globals, l echo.Logger,
) error {
	if len(name) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot remove name from network")
	}

	// TODO: quit with an error if the network was already named by DNS (disallow renaming)
	fqdn := fmt.Sprintf("%s.%s", name, cg.Config.DomainName)
	ttl := cg.Config.ZerotierDNS.NetworkTTL
	if _, err := client.CreateRRset(
		ctx, cg.DNSDomain, name, "TXT", ttl, []string{client.MakeNetworkIDRecord(id)}, l,
	); err != nil {
		// TODO: if a TXT RRset already exists but doesn't have the ID, just append a
		// zerotier-net-id=... record (but we should have a global lock on a get-and-patch
		// to avoid data races)
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't create a DNS TXT RRset at %s for network %s", fqdn, id,
		))
	}
	return client.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
}

func setNetworkRules(
	ctx context.Context, controller models.Controller, id string, jsonRules string,
) error {
	rules := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return err
	}

	err := client.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	if err != nil {
		return err
	}

	return nil
}

func postNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()
			l := c.Logger()

			// Parse params
			id := c.Param("id")
			address := client.GetControllerAddress(id)
			method := c.FormValue("method")

			// Run queries
			controller, err := client.FindControllerByAddress(
				ctx, address, app.Config, app.Cache, l,
			)
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
					ctx, *controller, id, c.FormValue("name"), app, l,
				); err != nil {
					return err
				}
			case "SETRULES":
				if err = setNetworkRules(
					ctx, *controller, id, c.FormValue("rules"),
				); err != nil {
					return err
				}

				return c.Redirect(
					http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id),
				)
			case "DELETE":
				if err = client.DeleteNetwork(ctx, *controller, id); err != nil {
					// TODO: add a tombstone to the TXT RRset?
					return err
				}

				return c.Redirect(http.StatusSeeOther, "/networks")
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
		}, nil
	}
}
