package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sargassum-eco/fluitans/pkg/framework/fingerprint"
	"github.com/sargassum-eco/fluitans/pkg/framework/httpcache"
)

type TemplateEtagSegments map[string][]string

func (te TemplateEtagSegments) newNotFoundError(t string) error {
	templates := make([]string, 0, len(te))
	for template := range te {
		templates = append(templates, template)
	}
	return fmt.Errorf(
		"couldn't find template etag segment for %s, as segments were only computed for [%s]",
		t, strings.Join(templates, ", "),
	)
}

func (te TemplateEtagSegments) Get(name string) ([]string, error) {
	tte, ok := te[name]
	if !ok {
		return nil, te.newNotFoundError(name)
	}
	return tte, nil
}

func (te TemplateEtagSegments) Require(names ...string) {
	for _, name := range names {
		_, ok := te[name]
		if !ok {
			panic(te.newNotFoundError(name))
		}
	}
}

func (te TemplateEtagSegments) SetAndCheckEtag(
	w http.ResponseWriter, r *http.Request, templateName string, data interface{},
) (noContent bool, err error) {
	// Look up data-independent etag segments
	templateEtagSegments, err := te.Get(templateName)
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
