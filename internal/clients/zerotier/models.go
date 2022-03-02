package zerotier

type ZTDNSSettings struct {
	NetworkTTL int64 `json:"networkTTL"`
	DeviceTTL  int64 `json:"deviceTTL"`
}
