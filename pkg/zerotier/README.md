# zerotier

Auto-generated code from the [ZeroTier One Service](https://docs.zerotier.com/service/v1/)'s OpenAPI specification.

This package provides Go bindings for writing HTTP clients to the ZeroTier One controller API. It forks ZeroTier's official OpenAPI spec file by:
- Adding a spec for DELETE `/controller/network/{networkID}`, since the DELETE method is specified for that path in Zerotier's [API documentation](github.com/zerotier/ZeroTierOne/blob/master/controller/README.md#controllernetworknetwork-id).
- Decomposing out the embedded objects from the schema definition for ControllerNetwork, to make the types easier to work with in Go
- Fixing a typo for the /controller/network/{networkID}/member/{nodeID} route, which was missing a `/` between `member` and `{nodeID}`
- Adding a spec for POST and DELETE of `/controller/network/{networkId}/member/{address}`, since those methods are specified for that path in ZeroTier's API documentation.

This package does not yet decompose out the embedded objects from any other schema definitions.

## Usage

To regenerate, make sure you've installed the [deepmap/oapi-codegen](github.com/deepmap/oapi-codegen) tool:
```
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

Then run `go generate ./...`.
