package networks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
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

type NetworkData struct {
	Controller       client.Controller
	Network          zerotier.ControllerNetwork
	Members          map[string]zerotier.ControllerNetworkMember
	JSONPrintedRules string
}

func getNetworkData(
	c echo.Context, address string, id string, cache *client.Cache,
) (*NetworkData, error) {
	controller, err := client.FindControllerByAddress(c, address, cache)
	if err != nil {
		return nil, err
	}

	network, memberAddresses, err := client.GetNetworkInfo(c, *controller, id)
	if err != nil {
		return nil, err
	}

	members, err := client.GetNetworkMembersInfo(c, *controller, id, memberAddresses)
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

	return &NetworkData{
		Controller:       *controller,
		Network:          *network,
		Members:          members,
		JSONPrintedRules: string(rules),
	}, nil
}

func getNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/network.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	switch cache := g.Cache.(type) {
	default:
		return nil, fmt.Errorf("global cache is of unexpected type %T", g.Cache)
	case *client.Cache:
		return func(c echo.Context) error {
			// Parse params
			id := c.Param("id")
			address := client.GetControllerAddress(id)

			// Run queries
			networkData, err := getNetworkData(c, address, id, cache)
			if err != nil {
				return err
			}

			// Handle Etag
			etagData, err := json.Marshal(networkData)
			if err != nil {
				return err
			}

			if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(etagData)); noContent {
				return err
			}

			// Render template
			return c.Render(http.StatusOK, t, struct {
				Meta   template.Meta
				Embeds template.Embeds
				Data   NetworkData
			}{
				Meta: template.Meta{
					Path:       c.Request().URL.Path,
					DomainName: client.GetDomainName(),
				},
				Embeds: g.Embeds,
				Data:   *networkData,
			})
		}, nil
	}
}

func renameNetwork(c echo.Context, controller client.Controller, id string, name string) error {
	var fqdn string
	if len(name) > 0 {
		fqdn = fmt.Sprintf("%s.%s", name, client.GetDomainName())
	} else {
		fqdn = ""
	}
	return client.UpdateNetwork(
		c, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
}

func setNetworkRules(
	c echo.Context, controller client.Controller, id string, jsonRules string,
) error {
	rules := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return err
	}

	err := client.UpdateNetwork(
		c, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	if err != nil {
		return err
	}

	return c.Redirect(
		http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id),
	)
}

func postNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch cache := g.Cache.(type) {
	default:
		return nil, fmt.Errorf("global cache is of unexpected type %T", g.Cache)
	case *client.Cache:
		return func(c echo.Context) error {
			// Parse params
			id := c.Param("id")
			address := client.GetControllerAddress(id)
			method := c.FormValue("method")

			// Run queries
			controller, err := client.FindControllerByAddress(c, address, cache)
			if err != nil {
				return err
			}

			switch method {
			case "RENAME":
				return renameNetwork(c, *controller, id, c.FormValue("name"))
			case "SETRULES":
				return setNetworkRules(c, *controller, id, c.FormValue("rules"))
			case "DELETE":
				err = client.DeleteNetwork(c, *controller, id)
				if err != nil {
					return err
				}

				return c.Redirect(http.StatusSeeOther, "/networks")
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
		}, nil
	}
}
