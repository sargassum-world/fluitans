// Package tmplfunc contains extension functions for templates
package tmplfunc

import (
	"html/template"
	"net/url"
)

type TurboStreamSigner func(streamName string) (signed string, err error)

func FuncMap(h HashedNamers, tss TurboStreamSigner) template.FuncMap {
	return template.FuncMap{
		"queryEscape":            url.QueryEscape,
		"identifyNetwork":        IdentifyNetwork,
		"getNetworkHostAddress":  GetNetworkHostAddress,
		"getNetworkNumber":       GetNetworkNumber,
		"durationToSec":          DurationToSec,
		"derefBool":              DerefBool,
		"derefInt":               DerefInt,
		"derefFloat32":           DerefFloat32,
		"derefString":            DerefString,
		"describeDNSRecordType":  DescribeDNSRecordType,
		"exemplifyDNSRecordType": ExemplifyDNSRecordType,
		"appHashed":              h.AppHashed,
		"staticHashed":           h.StaticHashed,
		"turboStreamSigned":      tss,
	}
}
