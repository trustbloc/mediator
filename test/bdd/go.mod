// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/test/bdd

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/cucumber/godog v0.9.0
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201009073544-dcb7f47ab8db
	github.com/trustbloc/edge-core v0.1.5-0.20200916124536-c32454a16108
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/hub-router => ../..
