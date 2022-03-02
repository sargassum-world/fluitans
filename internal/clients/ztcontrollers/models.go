package ztcontrollers

import (
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type Controller struct {
	Server            string  `json:"server"`
	Name              string  `json:"name"` // Must be unique for display purposes!
	Description       string  `json:"description"`
	Authtoken         string  `json:"authtoken"`
	NetworkCostWeight float32 `json:"local"`
}

func (c Controller) NewClient() (*zerotier.ClientWithResponses, error) {
	return zerotier.NewAuthClientWithResponses(c.Server, c.Authtoken)
}
