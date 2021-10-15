# Fluitans

Server application for provisioning and managing ZeroTier networks, domain names, organizations, and registries

Named after [_Sargassum fluitans_](https://www.algaebase.org/search/species/detail/?tc=accept&species_id=825) (the broadleaf gulfweed), one of the dominant sargassum species found in the Sargasso Sea, this server application is a prototype for a user-friendly way of managing ZeroTier networks.

## Usage

### Development

To run the server using golang's `go`, run `make run`. You will need to have installed golang first.

To execute the full build pipeline, run `make`; to build the docker images, run `make build`. You will need to have installed golang and golangci-lint first.

## License

Copyright Prakash Lab and the Sargassum project contributors.

SPDX-License-Identifier: Apache-2.0 OR BlueOak-1.0.0

You can use this project either under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0) or under the [Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0); you get to decide. We chose the Apache license because it's OSI-approved, and because it goes well together with the [Solderpad Hardware License](http://solderpad.org/licenses/SHL-2.1/), which is a license for open hardware used in other related projects but not this project. We prefer the Blue Oak Model License because it's easier to read and understand.
