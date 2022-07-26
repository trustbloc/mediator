// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/mediator/cmd/mediator

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.9-0.20220723135940-9bdda1afdad5
	github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb v0.0.0-20220428163625-96d8261511e1
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20220428211718-66cc046674a1
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20220606124520-53422361c38c
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/stretchr/testify v1.7.2
	github.com/trustbloc/edge-core v0.1.8
	github.com/trustbloc/mediator v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/mediator => ../..
