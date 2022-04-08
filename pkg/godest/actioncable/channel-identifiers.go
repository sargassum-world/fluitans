package actioncable

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type IdentifierChecker func(identifier string) error

// Identifier Parsers

func parseChannelName(identifier string) (channelName string, err error) {
	var i struct {
		Channel string `json:"channel"`
	}
	if err := json.Unmarshal([]byte(identifier), &i); err != nil {
		return "", errors.Wrap(err, "couldn't parse channel name from identifier")
	}
	return i.Channel, nil
}

func parseCSRFToken(identifier string) (token string, err error) {
	var i struct {
		Token string `json:"csrfToken"`
	}
	if err := json.Unmarshal([]byte(identifier), &i); err != nil {
		return "", errors.Wrap(err, "couldn't parse csrf token from identifier")
	}
	return i.Token, nil
}

func WithCSRFTokenChecker(checker func(token string) error) IdentifierChecker {
	return func(identifier string) error {
		token, err := parseCSRFToken(identifier)
		if err != nil {
			return err
		}
		if err := checker(token); err != nil {
			return errors.Wrap(err, "failed CSRF token check")
		}
		return nil
	}
}
