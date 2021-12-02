package templates

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// Asset hashed naming

type HashNamer func(string) string

func getHashedName(root string, namer HashNamer) HashNamer {
	return func(file string) string {
		return fmt.Sprintf("/%s/%s", root, namer(file))
	}
}

// HTTP error codes

type HTTPError struct {
	Name        string
	Description string
}

var httpErrors = map[int]HTTPError{
	http.StatusBadRequest: {
		Name:        "Bad request",
		Description: "The server cannot process the request due to something believed to be a client error.",
	},
	http.StatusUnauthorized: {
		Name:        "Unauthorized",
		Description: "The requested resource requires authentication.",
	},
	http.StatusForbidden: {
		Name:        "Access denied",
		Description: "Permission has not been granted to access the requested resource.",
	},
	http.StatusNotFound: {
		Name:        "Not found",
		Description: "The requested resource could not be found, but it may become available in the future.",
	},
	http.StatusTooManyRequests: {
		Name:        "Too busy",
		Description: "The server has reached a temporary usage limit. Please try again later.",
	},
	http.StatusInternalServerError: {
		Name:        "Server error",
		Description: "An unexpected problem occurred. We're working to fix it.",
	},
	http.StatusNotImplemented: {
		Name:        "Not implemented",
		Description: "The server cannot recognize the request method",
	},
	http.StatusBadGateway: {
		Name:        "Webservice currently unavailable",
		Description: "While handling the request, the server encountered a problem with another server. We're working to fix it.",
	},
	http.StatusServiceUnavailable: {
		Name:        "Webservice currently unavailable",
		Description: "The server is temporarily unable to handle the request. We're working to restore the server.",
	},
}

func describeError(code int) HTTPError {
	name, ok := httpErrors[code]
	if !ok {
		return HTTPError{
			Name:        "Server error",
			Description: "An unexpected problem occurred. We're working to fix it.",
		}
	}

	return name
}

// ZeroTier

func identifyNetwork(network zerotier.ControllerNetwork) string {
	if strings.TrimSpace(*network.Name) != "" {
		return *network.Name
	}

	return *network.Id
}

func getNetworkHostAddress(id string) string {
	return id[:10]
}

func getNetworkNumber(id string) string {
	return id[10:]
}

// DNS

type DNSRecordType struct {
	Description string
	Example     string
}

var recordTypes = map[string]DNSRecordType{
	"A": {
		Description: "IPv4 Addresses",
		Example:     "203.0.113.210",
	},
	"AAAA": {
		Description: "IPv6 Addresses",
		Example:     "2001:db8:2000:bf0::1",
	},
	"CAA": {
		Description: "Certificate Authorities",
		Example:     "0 issue \"ca.example.net\"",
	},
	"CERT": {
		Description: "Certificates",
		Example:     "",
	},
	"CNAME": {
		Description: "Canonical Name Aliases",
		Example:     "foo.example.com.",
	},
	"DNAME": {
		Description: "Name Delegations",
		Example:     "foo.example.com.",
	},
	"LOC": {
		Description: "Geographical Locations",
		Example:     "51 30 12.748 N 0 7 39.612 W 0.00",
	},
	"NS": {
		Description: "Delegated Authoritative Name Servers",
		Example:     "foo.example.com.",
	},
	"PTR": {
		Description: "Canonical Name Pointers",
		Example:     "foo.example.com.",
	},
	"RP": {
		Description: "Responsible Persons",
		Example:     "ethanli.stanford.edu ethanjli.people.fluitans.org",
	},
	"SRV": {
		Description: "Service Locations",
		Example:     "0 5 5060 sipserver.example.com",
	},
	"SSHFP": {
		Description: "SSH Public Host Key Fingerprints",
		Example:     "2 1 123456789abcdef67890123456789abcdef67890",
	},
	"TLSA": {
		Description: "TLS Server Certificates",
		Example:     "3 1 1 0123456789ABCDEF",
	},
	"TXT": {
		Description: "Textual Data",
		Example:     "\"zerotier-net-id=1c33c1ced015c144\"",
	},
	"URI": {
		Description: "URI Mappings",
		Example:     "10 1 \"ftp://ftp1.example.com/public\"",
	},
}

func describeDNSRecordType(recordType string) string {
	recordTypeInfo, ok := recordTypes[recordType]
	if !ok {
		return "Unknown Record Type"
	}

	return recordTypeInfo.Description
}

func exemplifyDNSRecordType(recordType string) string {
	recordTypeInfo, ok := recordTypes[recordType]
	if !ok {
		return ""
	}

	return recordTypeInfo.Example
}

// Durations

func durationToSec(i time.Duration) float64 {
	return i.Seconds()
}

// Pointers

func derefBool(b *bool) bool {
	return b != nil && *b
}

func derefInt(i *int, nilValue int) int {
	if i == nil {
		return nilValue
	}

	return *i
}

func derefFloat32(i *float32, nilValue float32) float32 {
	if i == nil {
		return nilValue
	}

	return *i
}

func derefString(s *string, nilValue string) string {
	if s == nil {
		return nilValue
	}

	return *s
}
