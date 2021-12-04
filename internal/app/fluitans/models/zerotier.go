package models

type Controller struct {
	Server            string  `json:"server"`
	Name              string  `json:"name"` // Must be unique for display purposes!
	Description       string  `json:"description"`
	Authtoken         string  `json:"authtoken"`
	NetworkCostWeight float32 `json:"local"`
}

type ZerotierDNSSettings struct {
	NetworkTTL int64 `json:"networkTTL"`
	DeviceTTL  int64 `json:"deviceTTL"`
}
