// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/mediator/cmd/mediator

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.9-0.20220726210616-c28f931361fe
	github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb v0.0.0-20220615170242-cda5092b4faf
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20220428211718-66cc046674a1
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20220614152730-3d817acfa48b
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/stretchr/testify v1.7.2
	github.com/trustbloc/edge-core v0.1.8
	github.com/trustbloc/mediator v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/mediator => ../..
