package clientcache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type Marshaller interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(marshaled []byte, result interface{}) error
}

// Gob

type GobMarshaller struct { // The Gob marshaller would be faster if it reused the encoder and decoder, instead of
	// constructing new ones on each method call. However, then it wouldn't be concurrency-safe.
	// We prefer to use MsgPack because it's probably faster anyways.
}

func NewGobMarshaller() GobMarshaller {
	return GobMarshaller{}
}

func (m *GobMarshaller) Marshal(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't gob-encode value %#v", value))
	}
	return buf.Bytes(), nil
}

func (m *GobMarshaller) Unmarshal(marshaled []byte, result interface{}) error {
	buf := bytes.NewBuffer(marshaled)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(result); err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't gob-decode type %T from bytes %+v", result, marshaled,
		))
	}
	return nil
}

// MsgPack

type MsgPackMarshaller struct{}

func NewMsgPackMarshaller() MsgPackMarshaller {
	return MsgPackMarshaller{}
}

func (m *MsgPackMarshaller) Marshal(value interface{}) ([]byte, error) {
	marshaled, err := msgpack.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't msgpack-encode value %#v", value))
	}
	return marshaled, nil
}

func (m *MsgPackMarshaller) Unmarshal(marshaled []byte, result interface{}) error {
	if err := msgpack.Unmarshal(marshaled, result); err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't msgpack-decode type %T from bytes %+v", result, marshaled,
		))
	}
	return nil
}
