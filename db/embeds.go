// Package db contains application-specific db schemas and scripts
package db

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/godest/database"
	sessions "github.com/sargassum-world/godest/session/sqlitestore"
)

// Randomly-generated 32-bit integer for the fluitans app, to prevent accidental migration of database
// files from other applications.
const appID = 250256223

// Migrations

var DomainEmbeds map[string]database.DomainEmbeds = map[string]database.DomainEmbeds{
	"sessions": sessions.NewDomainEmbeds(),
}

var MigrationFiles []database.MigrationFile = []database.MigrationFile{
	{Domain: "sessions", File: sessions.MigrationFiles[0]},
}

// Queries

var (
	//go:embed queries/*
	queriesEFS              embed.FS
	queriesFS, _            = fs.Sub(queriesEFS, "queries")
	prepareConnQueriesFS, _ = fs.Sub(queriesFS, "prepare-conn")
)

// Embeds

func NewEmbeds() database.Embeds {
	return database.Embeds{
		AppID:                appID,
		DomainEmbeds:         DomainEmbeds,
		MigrationFiles:       MigrationFiles,
		PrepareConnQueriesFS: prepareConnQueriesFS,
	}
}
