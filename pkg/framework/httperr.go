package framework

import (
	"net/http"
)

type HTTPError struct {
	Name        string
	Description string
}

var HTTPErrors = map[int]HTTPError{
	http.StatusBadRequest: {
		Name:        "Bad request",
		Description: "The server cannot process the request due to something believed to be a client error.",
	},
	http.StatusUnauthorized: {
		Name:        "Unauthenticated",
		Description: "The requested resource requires authentication.",
	},
	http.StatusForbidden: {
		Name:        "Access denied",
		Description: "Permission has not been granted to access the requested resource.",
	},
	http.StatusNotFound: {
		Name: "Not available",
		Description: "The requested resource is not available because it could not be found, " +
			"it requires authentication, or permission has not been granted to access it.",
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
		Description: "The server cannot recognize the request method.",
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

func DescribeHTTPError(code int) HTTPError {
	name, ok := HTTPErrors[code]
	if !ok {
		return HTTPError{
			Name:        "Server error",
			Description: "An unexpected problem occurred. We're working to fix it.",
		}
	}

	return name
}
