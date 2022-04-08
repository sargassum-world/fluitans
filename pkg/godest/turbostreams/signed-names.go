package turbostreams

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
)

type Signer struct {
	Config SignerConfig
}

func NewSigner(config SignerConfig) Signer {
	return Signer{
		Config: config,
	}
}

func (s Signer) Check(identifier string) error {
	name, err := s.parseIdentifier(identifier)
	if err != nil {
		return err
	}
	if !s.validate(name) {
		return errors.Errorf("signed stream name %s failed HMAC check", name.Name)
	}
	return nil
}

type signedName struct {
	Name string
	Hash []byte
}

func (s Signer) validate(n signedName) bool {
	return hmac.Equal(n.Hash, s.hash(n.Name))
}

func (s Signer) parseIdentifier(identifier string) (parsed signedName, err error) {
	var params struct {
		Name string `json:"name"`
		Hash string `json:"integrity"`
	}
	if err = json.Unmarshal([]byte(identifier), &params); err != nil {
		return signedName{}, errors.Wrap(err, "couldn't parse identifier for params")
	}
	parsed.Name = params.Name
	parsed.Hash, err = base64.StdEncoding.DecodeString(params.Hash)
	if err != nil {
		return signedName{}, errors.Wrap(err, "couldn't base64-decode stream name hash")
	}
	return parsed, nil
}

func (s Signer) hash(streamName string) []byte {
	h := hmac.New(sha512.New, s.Config.HashKey)
	h.Write([]byte(streamName))
	return h.Sum(nil)
}

func (s Signer) Sign(streamName string) (hash string) {
	return base64.StdEncoding.EncodeToString(s.hash(streamName))
}
