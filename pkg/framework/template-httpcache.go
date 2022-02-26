package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fingerprint"
	"github.com/sargassum-eco/fluitans/pkg/framework/httpcache"
)

type Fingerprints struct {
	App  string
	Page map[string]string
}

func (f Fingerprints) GetEtagSegments(templateName string) ([]string, error) {
	if templateName == "" {
		return []string{f.App}, nil
	}

	pageFingerprint, ok := f.Page[templateName]
	if !ok {
		return []string{f.App}, errors.Errorf(
			"couldn't find page fingerprint for template %s", templateName,
		)
	}

	return []string{f.App, pageFingerprint}, nil
}

func (f Fingerprints) MustHave(templateNames ...string) {
	for _, name := range templateNames {
		if _, err := f.GetEtagSegments(name); err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("couldn't find template etag segments for %s", name)))
		}
	}
}

func (f Fingerprints) SetAndCheckEtag(
	w http.ResponseWriter, r *http.Request, templateName string, data interface{},
) (noContent bool, err error) {
	// Look up data-independent etag segments
	templateEtagSegments, err := f.GetEtagSegments(templateName)
	if err != nil {
		return
	}

	// Encode data
	var buf bytes.Buffer
	// github.com/vmihailenco/msgpack has better performance, but we use the JSON encoder because
	// the msgpack encoder can only sort the map keys of map[string]string and map[string]interface{}
	// maps, and it's too much trouble to convert our maps into map[string]interface{}. If we can
	// work around this limitation, we should use msgpack though.
	if err = json.NewEncoder(&buf).Encode(data); err != nil {
		return
	}
	encoded := buf.Bytes()

	noContent = httpcache.SetAndCheckEtag(
		w, r, append(templateEtagSegments, fingerprint.Compute(encoded))...,
	)
	return
}
