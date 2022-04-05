// Package turbostreams provides server-side support for sending Hotwired Turbo Streams in POST
// responses as well as over Action Cables.
package turbostreams

import (
	_ "embed"
)

type Action string

const (
	ActionAppend  Action = "append"
	ActionPrepend Action = "prepend"
	ActionReplace Action = "replace"
	ActionUpdate  Action = "update"
	ActionRemove  Action = "remove"
	ActionBefore  Action = "before"
	ActionAfter   Action = "after"
)

type Message struct {
	Action   Action
	Target   string
	Targets  string
	Template string
	Data     interface{}
}

//go:embed streams.partial.tmpl
var Template string
