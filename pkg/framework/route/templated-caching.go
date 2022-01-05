package route

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

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
	c echo.Context, templateEtagSegments []string, data interface{},
) (bool, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	enc.SetSortMapKeys(true)
	if err := enc.Encode(data); err != nil {
		return false, err
	}
	return httpcache.ProcessEtag(
		c, append(templateEtagSegments, fingerprint.Compute(buf.Bytes()))...,
	)
}
