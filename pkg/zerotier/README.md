# zerotier

Auto-generated code from the [ZeroTier One Service](https://docs.zerotier.com/service/v1/)'s OpenAPI specification.

This package provides Go bindings for writing HTTP clients to the ZeroTier One controller API. It forks ZeroTier's official OpenAPI spec file by:
- Adding a spec for DELETE `/controller/network/{networkID}`, since the DELETE method is specified for that path in Zerotier's [API documentation](github.com/zerotier/ZeroTierOne/blob/master/controller/README.md#controllernetworknetwork-id).
- Decomposing out the embedded objects from the schema definition for ControllerNetwork, to make the types easier to work with in Go

This package does not yet decompose out the embedded objects from any other schema definitions.

## Usage

To regenerate, install and run the [deepmap/oapi-codegen](github.com/deepmap/oapi-codegen) tool:
```
go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
oapi-codegen --generate=types zerotier.json > types.gen.go
oapi-codegen --generate=client zerotier.json > client.gen.go
oapi-codegen --generate=server zerotier.json > server.gen.go
```
Then you will have to manually change the package name in the generated files from `Zerotier` to `zerotier`.
