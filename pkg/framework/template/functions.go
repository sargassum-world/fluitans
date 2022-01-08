package template

import (
	"fmt"
	"html/template"
	"net/http"
)

// All functions

func FuncMap(appNamer, staticNamer HashNamer) template.FuncMap {
	return template.FuncMap{
		"appHashed":     getHashedName("app", appNamer),
		"staticHashed":  getHashedName("static", staticNamer),
		"describeError": describeError,
	}
}

// Asset hashed naming

type HashNamer func(string) string

func getHashedName(root string, namer HashNamer) HashNamer {
	return func(file string) string {
		return fmt.Sprintf("/%s/%s", root, namer(file))
	}
}

// HTTP error codes

type HTTPError struct {
	Name        string
	Description string
}

var httpErrors = map[int]HTTPError{
	http.StatusBadRequest: {
		Name:        "Bad request",
		Description: "The server cannot process the request due to something believed to be a client error.",
	},
	http.StatusUnauthorized: {
		Name:        "Unauthorized",
		Description: "The requested resource requires authentication.",
	},
	http.StatusForbidden: {
		Name:        "Access denied",
		Description: "Permission has not been granted to access the requested resource.",
	},
	http.StatusNotFound: {
		Name:        "Not found",
		Description: "The requested resource could not be found, but it may become available in the future.",
	},
	http.StatusTooManyRequests: {
		Name:        "Too busy",
		Description: "The server has reached a temporary usage limit. Please try again later.",
	},
	http.StatusInternalServerError: {
		Name:        "Server error",
		Description: "An unexpected problem occurred. We're working to fix it.",
	},
	http.StatusNotImplemented: {
		Name:        "Not implemented",
		Description: "The server cannot recognize the request method",
	},
	http.StatusBadGateway: {
		Name: "Webservice currently unavailable",
		Description: "While handling the request, the server encountered a problem with another server. " +
			"We're working to fix it.",
	},
	http.StatusServiceUnavailable: {
		Name: "Webservice currently unavailable",
		Description: "The server is temporarily unable to handle the request. " +
			"We're working to restore the server.",
	},
}

func describeError(code int) HTTPError {
	name, ok := httpErrors[code]
	if !ok {
		return HTTPError{
			Name:        "Server error",
			Description: "An unexpected problem occurred. We're working to fix it.",
		}
	}

	return name
}