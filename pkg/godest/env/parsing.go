// Package env contains code for handling environment variables
package env

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

func GetBool(varName string) (bool, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return false, nil
	}

	switch value {
	case "TRUE", "true", "True":
		return true, nil
	case "FALSE", "false", "False":
		return false, nil
	}

	return false, errors.Errorf(
		"unknown value %s for boolean environment variable %s", value, varName,
	)
}

func GetUint64(varName string, defaultValue uint64) (uint64, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue, nil
	}

	const (
		base  = 10
		width = 32 // bits
	)
	parsed, err := strconv.ParseUint(value, base, width)
	if err != nil {
		return 0, errors.Wrapf(
			err, "unparseable value %s for uint64 environment variable %s", value, varName,
		)
	}

	return parsed, nil
}

func GetInt64(varName string, defaultValue int64) (int64, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue, nil
	}

	const (
		base  = 10
		width = 64 // bits
	)
	parsed, err := strconv.ParseInt(value, base, width)
	if err != nil {
		return 0, errors.Wrapf(
			err, "unparseable value %s for int64 environment variable %s", value, varName,
		)
	}

	return parsed, nil
}

func GetFloat32(varName string, defaultValue float32) (float32, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue, nil
	}

	const width = 32 // bits
	parsed, err := strconv.ParseFloat(value, width)
	if err != nil {
		return 0, errors.Wrapf(
			err, "unparseable value %s for float32 environment variable %s", value, varName,
		)
	}

	return float32(parsed), nil
}

func GetString(varName string, defaultValue string) string {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue
	}

	return value
}

func GetBase64(varName string) ([]byte, error) {
	rawValue := os.Getenv(varName)
	if len(rawValue) == 0 {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(rawValue)
}

func GetURL(varName string, defaultValue string) (*url.URL, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		value = defaultValue
	}

	return url.Parse(value)
}

func GetURLOrigin(varName, defaultValue, defaultScheme string) (*url.URL, error) {
	url, err := GetURL(varName, defaultValue)
	if err != nil {
		return nil, errors.Wrapf(
			err, "unparseable value %s for URL environment variable %s", os.Getenv(varName), varName,
		)
	}

	if len(url.Scheme) == 0 {
		url.Scheme = defaultScheme
	}
	url.Path = ""
	url.User = nil
	url.RawQuery = ""
	url.Fragment = ""

	return url, nil
}

// GenerateRandomKey creates a random key with the given length in bytes.
// On failure, returns nil.
//
// Note that keys created using `GenerateRandomKey()` are not automatically
// persisted. New keys will be created when the application is restarted, and
// previously issued cookies will not be able to be decoded.
//
// Callers should explicitly check for the possibility of a nil return, treat
// it as a failure of the system random number generator, and not continue.
func GenerateRandomKey(length int) []byte {
	// Note: copied from github.com/gorilla/securecookie
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

func GetKey(varName string, length int) ([]byte, error) {
	key, err := GetBase64(varName)
	if err != nil {
		return nil, err
	}

	if key == nil {
		hashKeySize := 32
		key = GenerateRandomKey(hashKeySize)
		if key == nil {
			return nil, errors.New("unable to generate a random key")
		}
		// TODO: print to the logger instead?
		fmt.Printf(
			"Record this key for future use as %s: %s\n",
			varName, base64.StdEncoding.EncodeToString(key),
		)
	}
	return key, nil
}
