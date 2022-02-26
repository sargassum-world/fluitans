// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"time"

	"github.com/sargassum-eco/fluitans/pkg/godest"
)

const (
	AppURLPrefix    = "/app/"
	StaticURLPrefix = "/static/"
	FontsURLPrefix  = "/fonts/"
)

type TemplatedService struct {
	r godest.TemplateRenderer
}

func NewTemplatedService(r godest.TemplateRenderer) *TemplatedService {
	return &TemplatedService{
		r: r,
	}
}

func (s *TemplatedService) Register(er godest.EchoRouter) {
	er.GET(AppURLPrefix+"app.webmanifest", s.getWebmanifest())
}

func RegisterStatic(er godest.EchoRouter, em godest.Embeds) {
	const (
		day  = 24 * time.Hour
		week = 7 * day
		year = 365 * day
	)

	er.GET("/favicon.ico", godest.HandleFS("/", em.StaticFS, week))
	er.GET(FontsURLPrefix+"*", godest.HandleFS(FontsURLPrefix, em.FontsFS, year))
	er.GET(StaticURLPrefix+"*", godest.HandleFSFileRevved(StaticURLPrefix, em.StaticHFS))
	er.GET(AppURLPrefix+"*", godest.HandleFSFileRevved(AppURLPrefix, em.AppHFS))
}
