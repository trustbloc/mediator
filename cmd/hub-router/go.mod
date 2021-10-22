// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/cmd/hub-router

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.8-0.20211020135731-5ab9da9e29ab
	github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb v0.0.0-20210913191140-6a8013cdda32
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210813123233-e22ddceee0b1
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20211020135731-5ab9da9e29ab
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20211020135731-5ab9da9e29ab
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/edge-core v0.1.7
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/hub-router => ../..
