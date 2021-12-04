package tmplfunc

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
