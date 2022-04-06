package turbostreams

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "TURBOSTREAMS_"

type SignerConfig struct {
	HashKey []byte
}

func GetSignerConfig() (c SignerConfig, err error) {
	const hashKeySize = 32
	c.HashKey, err = env.GetKey(envPrefix+"HASH_KEY", hashKeySize)
	if err != nil {
		return SignerConfig{}, errors.Wrap(err, "couldn't make turbo streams name signer config")
	}

	return c, nil
}
