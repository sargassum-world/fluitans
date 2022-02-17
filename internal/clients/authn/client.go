// Package authn provides a high-level client for user authentication
package authn

import (
	"crypto/subtle"

	"github.com/alexedwards/argon2id"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

type Client struct {
	Config Config
	Logger log.Logger
}

func MakeClient(l log.Logger) (*Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions client config")
	}

	return &Client{
		Config: *config,
		Logger: l,
	}, nil
}

func (c *Client) CheckCredentials(username, password string) (bool, error) {
	if c.Config.NoAuth {
		return true, nil
	}

	usernameBytes := []byte(username)
	adminUsername := []byte(c.Config.AdminUsername)
	usernameMatch := subtle.ConstantTimeCompare(usernameBytes, adminUsername) == 1

	// TODO: if the username doesn't match, can we safely skip checking the password without leaking
	// timing information about the validity of a username? e.g. can we measure how long the password
	// comparison takes and just sleep for that duration?
	passwordMatch, err := argon2id.ComparePasswordAndHash(password, c.Config.AdminPasswordHash)
	if err != nil {
		return false, err
	}
	return usernameMatch && passwordMatch, nil
}
