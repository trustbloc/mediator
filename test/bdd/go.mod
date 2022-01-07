// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/test/bdd

go 1.17

require (
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/cucumber/godog v0.9.0
	github.com/fsouza/go-dockerclient v1.6.6
	github.com/google/uuid v1.3.0
	github.com/hyperledger/aries-framework-go v0.1.8-0.20211231170827-1f7d634dfcec
	github.com/hyperledger/aries-framework-go/test/bdd v0.0.0-20211215163235-9b915aa86dbe
	github.com/trustbloc/edge-core v0.1.7
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
	gotest.tools/v3 v3.0.3 // indirect
)

replace (
	github.com/trustbloc/hub-router => ../..
	golang.org/x/sys => golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
)
