// Package env contains code for handling environment variables
package env

import (
	"fmt"
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

func GetInt64(varName string, defaultValue int64) (int64, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue, nil
	}

	base := 10
	width := 64 // bits
	parsed, err := strconv.ParseInt(value, base, width)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf(
			"unparseable value %s for int64 environment variable %s", value, varName,
		))
	}

	return parsed, nil
}

func GetFloat32(varName string, defaultValue float32) (float32, error) {
	value := os.Getenv(varName)
	if len(value) == 0 {
		return defaultValue, nil
	}

	width := 32 // bits
	parsed, err := strconv.ParseFloat(value, width)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf(
			"unparseable value %s for float32 environment variable %s", value, varName,
		))
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
		return nil, errors.Wrap(err, fmt.Sprintf(
			"unparseable value %s for URL environment variable %s", os.Getenv(varName), varName,
		))
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
