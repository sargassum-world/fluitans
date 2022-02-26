// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"time"

	"github.com/sargassum-eco/fluitans/pkg/framework"
)

const (
	AppURLPrefix    = "/app/"
	StaticURLPrefix = "/static/"
	FontsURLPrefix  = "/fonts/"
)

type TemplatedService struct {
	r framework.TemplateRenderer
}

func NewTemplatedService(r framework.TemplateRenderer) *TemplatedService {
	return &TemplatedService{
		r: r,
	}
}

func (s *TemplatedService) Register(er framework.EchoRouter) {
	er.GET(AppURLPrefix+"app.webmanifest", s.getWebmanifest())
}

func RegisterStatic(er framework.EchoRouter, em framework.Embeds) {
	const (
		day  = 24 * time.Hour
		week = 7 * day
		year = 365 * day
	)

	er.GET("/favicon.ico", framework.HandleFS("/", em.StaticFS, week))
	er.GET(FontsURLPrefix+"*", framework.HandleFS(FontsURLPrefix, em.FontsFS, year))
	er.GET(StaticURLPrefix+"*", framework.HandleFSFileRevved(StaticURLPrefix, em.StaticHFS))
	er.GET(AppURLPrefix+"*", framework.HandleFSFileRevved(AppURLPrefix, em.AppHFS))
}
