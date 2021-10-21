package zerotier

import (
	"context"
	"net/http"
)

// NewAuthClient creates a new Client, with the server's required authtoken and reasonable defaults.
func NewAuthClient(server string, authtoken string, opts ...ClientOption) (*Client, error) {
	return NewClient(
		server,
		append(
			[]ClientOption{WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
				req.Header.Set("X-ZT1-Auth", authtoken)
				return nil
			})},
			opts...,
		)...,
	)
}

// NewAuthClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling and the server's required authtoken.
func NewAuthClientWithResponses(server string, authtoken string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewAuthClient(server, authtoken, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}
