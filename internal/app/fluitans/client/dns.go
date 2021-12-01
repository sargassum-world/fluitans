package client

type DNSServer struct {
	// TODO: move this to a models package or something!
	Server            string  `json:"server"`
	API               string  `json:"api"`
	Name              string  `json:"name"` // Must be unique for display purposes!
	Description       string  `json:"description"`
	Authtoken         string  `json:"authtoken"`
	NetworkCostWeight float32 `json:"local"`
}

var RecordTypes []string = []string{
	"A",
	"AAAA",
	// "CAA",
	// "CERT",
	"CNAME",
	"DNAME",
	"LOC",
	"NS",
	"PTR",
	"RP",
	"SRV",
	// "SSHFP",
	// "TLSA",
	"TXT",
	"URI",
}
