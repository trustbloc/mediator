// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/test/bdd

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/cucumber/godog v0.9.0
	github.com/docker/docker v1.4.2-0.20200319182547-c7ad2b866182 // indirect
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210427144858-06fb8b7d2d30
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210429200350-4099d2551ddd // indirect
	github.com/trustbloc/edge-core v0.1.7-0.20210426154540-f9c761ec6943
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
	gotest.tools/v3 v3.0.3 // indirect
)

replace (
	github.com/trustbloc/hub-router => ../..
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
)
