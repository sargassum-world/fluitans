package godest

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type fingerprints struct {
	app  string
	page map[string]string
}

func (f fingerprints) getEtagSegments(templateName string) ([]string, error) {
	if templateName == "" {
		return []string{f.app}, nil
	}

	pageFingerprint, ok := f.page[templateName]
	if !ok {
		return []string{f.app}, errors.Errorf(
			"couldn't find page fingerprint for template %s", templateName,
		)
	}

	return []string{f.app, pageFingerprint}, nil
}

func (f fingerprints) mustHaveForPage(templateName string) {
	if _, err := f.getEtagSegments(templateName); err != nil {
		panic(errors.Wrapf(
			err, "couldn't find template etag segments for page template %s", templateName,
		))
	}
}

func (f fingerprints) setAndCheckEtag(
	w http.ResponseWriter, r *http.Request, templateName string, data interface{},
) (noContent bool, err error) {
	// Look up data-independent etag segments
	templateEtagSegments, err := f.getEtagSegments(templateName)
	if err != nil {
		return false, err
	}

	// Encode data
	var buf bytes.Buffer
	// github.com/vmihailenco/msgpack has better performance, but we use the JSON encoder because
	// the msgpack encoder can only sort the map keys of map[string]string and map[string]interface{}
	// maps, and it's too much trouble to convert our maps into map[string]interface{}. If we can
	// work around this limitation, we should use msgpack though.
	if err = json.NewEncoder(&buf).Encode(data); err != nil {
		return false, err
	}
	encoded := buf.Bytes()

	noContent = setAndCheckEtag(
		w, r, append(templateEtagSegments, computeFingerprint(encoded))...,
	)
	if noContent {
		w.WriteHeader(http.StatusNotModified)
	}
	return noContent, nil
}
