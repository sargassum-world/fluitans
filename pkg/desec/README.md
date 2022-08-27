# desec

Auto-generated code from the [deSEC DNS API](https://desec.readthedocs.io/en/latest/index.html).

This package provides Go bindings for writing HTTP clients to the deSEC DNS management API. It forks the [draft OpenAPI specification](https://github.com/desec-io/desec-stack/issues/359#issuecomment-865365725) spec file by:
- Replacing `\.` with `.` in the `/{name}/rrsets/{subname}.../{type}/` route so that the Go code generates correcly instead of failing with an "invalid escape sequence" error.
- Disabling generation of client code for the `/api/v1/domains/{name}/rrsets/{subname}/{type}/` and `/api/v1/domains/{name}/rrsets/{subname}@/{type}/`, routes, as those routes are redundant to the `/api/v1/domains/{name}/rrsets/{subname}.../{type}/` route and we don't need multiple sets of client code functions (which would requie different names) to do the same thing.
- Disabling generation of client code for the `/api/v1/domains/{name}/rrsets/.../{type}/` and `/api/v1/domains/{name}/rrsets/@/{type}/` routes, as those routes are redundant to the `/api/v1/domains/{name}/rrsets/{subname}.../{type}/` (with an empty string for the `subname` URL parameter) route.
- Adding a Key component to the list of models, as the Domain component now has a list of Key objects rather than a list of strings for its `keys` field.
- Added the `subname` and `type` GET query parameters to the `/api/v1/domains/{name}/rrsets/` route.
- Updated the schemas for expected responses from the `/api/v1/domains/{name}/rrsets/` and `/api/v1/domains/` routes to be an array of RRset objects, rather than an object containing that array along with pagination cursors.
- Updated the RRset object schema to make records be an array of strings, rather than an array of objects each containing a `content` field.

## Usage

To regenerate, make sure you've installed the [deepmap/oapi-codegen](github.com/deepmap/oapi-codegen) tool:
```
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

Then run `go generate ./...`.
