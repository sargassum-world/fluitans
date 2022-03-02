// Package models provides shared data models
package models

import (
	"github.com/sargassum-eco/fluitans/pkg/desec"
)

type DNSServer struct {
	Server            string  `json:"server"`
	API               string  `json:"api"`
	Name              string  `json:"name"` // Must be unique for display purposes!
	Description       string  `json:"description"`
	Authtoken         string  `json:"authtoken"`
	NetworkCostWeight float32 `json:"local"`
}

func (s DNSServer) NewClient() (*desec.ClientWithResponses, error) {
	return desec.NewAuthClientWithResponses(s.Server, s.Authtoken)
}
