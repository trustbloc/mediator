// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/cmd/hub-router

go 1.15

require (
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210301183320-85351acdb748
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210224230531-58e1368e5661
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210302153503-0e00e248f14d
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/edge-core v0.1.6-0.20210212172534-81ab3a5abf5b
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/hub-router => ../..
