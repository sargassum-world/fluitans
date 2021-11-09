package client

type Controller struct {
	// TODO: move this to a models package or something!
	Server            string  `json:"server"`
	Name              string  `json:"name"` // Must be unique for display purposes!
	Description       string  `json:"description"`
	Authtoken         string  `json:"authtoken"`
	NetworkCostWeight float32 `json:"local"`
}

func GetControllers() ([]Controller, error) {
	// TODO: look up the controllers from a database, if one is specified!
	controllers := make([]Controller, 0)
	envController, err := GetEnvVarController()
	if err != nil {
		return nil, err
	}

	if envController != nil {
		controllers = append(controllers, *envController)
	}
	return controllers, nil
}

func FindController(name string) (*Controller, bool, error) {
	found := false
	controllers, err := GetControllers()
	if err != nil {
		return nil, false, err
	}

	for _, v := range controllers {
		if v.Name == name {
			return &v, true, nil
		}
	}
	return nil, found, nil
}
