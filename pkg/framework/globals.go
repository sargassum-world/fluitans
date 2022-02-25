// Package framework provides a reusable framework for Fluitans-style web apps
package framework

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/embeds"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Globals struct {
	Template route.TemplateGlobals
}

func NewGlobals(e embeds.Embeds) (*Globals, error) {
	af, err := e.ComputeAppFingerprint()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute fingerprint for app")
	}

	pf, err := e.ComputePageFingerprints()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute fingerprint for page/module templates")
	}

	return &Globals{
		Template: route.TemplateGlobals{
			Inlines:          e.Inlines,
			AppFingerprint:   af,
			PageFingerprints: pf,
		},
	}, nil
}
