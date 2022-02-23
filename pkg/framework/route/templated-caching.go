package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fingerprint"
	"github.com/sargassum-eco/fluitans/pkg/framework/httpcache"
)

type TemplateEtagSegments map[string][]string

func (te TemplateEtagSegments) GetSegment(name string) ([]string, error) {
	tte, ok := te[name]
	if !ok {
		return nil, te.NewNotFoundError(name)
	}
	return tte, nil
}

func (te TemplateEtagSegments) RequireSegments(route string, names ...string) error {
	for _, name := range names {
		_, ok := te[name]
		if !ok {
			return errors.Wrap(te.NewNotFoundError(name), fmt.Sprintf(
				"couldn't find template etag segment for %s", route,
			))
		}
	}
	return nil
}

func (te TemplateEtagSegments) NewNotFoundError(t string) error {
	templates := make([]string, 0, len(te))
	for template := range te {
		templates = append(templates, template)
	}
	return fmt.Errorf(
		"couldn't find template etag segment for %s in [%s]",
		t, strings.Join(templates, ", "),
	)
}

func ProcessEtag(
	w http.ResponseWriter, r *http.Request, templateEtagSegments []string, data interface{},
) (bool, error) {
	var buf bytes.Buffer
	// github.com/vmihailenco/msgpack has better performance, but we use the JSON encoder because
	// the msgpack encoder can only sort the map keys of map[string]string and map[string]interface{}
	// maps, and it's too much trouble to convert our maps into map[string]interface{}. If we can
	// work around this limitation, we should use msgpack though.
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return false, err
	}
	encoded := buf.Bytes()
	return httpcache.ProcessEtag(
		w, r, append(templateEtagSegments, fingerprint.Compute(encoded))...,
	), nil
}
