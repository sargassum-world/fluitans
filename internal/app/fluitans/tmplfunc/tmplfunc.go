// Package tmplfunc contains extension functions for templates
package tmplfunc

import (
	"html/template"
)

var FuncMap template.FuncMap = template.FuncMap{
	"identifyNetwork":        identifyNetwork,
	"getNetworkHostAddress":  getNetworkHostAddress,
	"getNetworkNumber":       getNetworkNumber,
	"durationToSec":          durationToSec,
	"derefBool":              derefBool,
	"derefInt":               derefInt,
	"derefFloat32":           derefFloat32,
	"derefString":            derefString,
	"describeDNSRecordType":  describeDNSRecordType,
	"exemplifyDNSRecordType": exemplifyDNSRecordType,
}
