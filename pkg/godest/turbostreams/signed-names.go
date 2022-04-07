package turbostreams

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
)

type Signer struct {
	Config SignerConfig
}

func NewSigner(config SignerConfig) Signer {
	return Signer{
		Config: config,
	}
}

func (s Signer) NewChannel(
	identifier string, h *MessagesHub, handleSub SubHandler, handleMsg MsgHandler,
) (*Channel, error) {
	name, err := s.parseIdentifier(identifier)
	if err != nil {
		return nil, err
	}
	if !s.validate(name) {
		return nil, errors.Errorf("signed stream name %s failed HMAC check", name.Name)
	}
	return &Channel{
		identifier: identifier,
		name:       name,
		h:          h,
		handleSub:  handleSub,
		handleMsg:  handleMsg,
	}, nil
}

func (s Signer) ChannelFactory(
	h *MessagesHub, handleSub SubHandler, handleMsg MsgHandler,
) actioncable.ChannelFactory {
	return func(identifier string) (actioncable.Channel, error) {
		return s.NewChannel(identifier, h, handleSub, handleMsg)
	}
}

type signedName struct {
	Name string `msgpack:"name"`
	Hash []byte `msgpack:"hash"`
}

func (s Signer) validate(n signedName) bool {
	return hmac.Equal(n.Hash, s.hash(n.Name))
}

func (s Signer) parseIdentifier(identifier string) (signedName, error) {
	var p struct {
		SignedName string `json:"signed_stream_name"`
	}
	if err := json.Unmarshal([]byte(identifier), &p); err != nil {
		return signedName{}, errors.Wrap(err, "couldn't parse identifier for params")
	}
	signedRaw, err := base64.StdEncoding.DecodeString(p.SignedName)
	if err != nil {
		return signedName{}, errors.Wrap(err, "couldn't base64-decode signed stream name")
	}
	var name signedName
	err = msgpack.Unmarshal(signedRaw, &name)
	return name, errors.Wrap(err, "couldn't parse name and hash from decoded signed stream name")
}

func (s Signer) hash(streamName string) []byte {
	h := hmac.New(sha512.New, s.Config.HashKey)
	h.Write([]byte(streamName))
	return h.Sum(nil)
}

func (s Signer) Sign(streamName string) (signed string, err error) {
	signedRaw, err := msgpack.Marshal(signedName{
		Name: streamName,
		Hash: s.hash(streamName),
	})
	if err != nil {
		return "", errors.Wrap(err, "couldn't marshal stream name and hash")
	}
	return base64.StdEncoding.EncodeToString(signedRaw), nil
}
