// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-router/test/bdd

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/cucumber/godog v0.9.0
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/google/uuid v1.1.2
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210302153503-0e00e248f14d
	github.com/opencontainers/runc v1.0.0-rc9 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/trustbloc/edge-core v0.1.6-0.20210212172534-81ab3a5abf5b
	github.com/trustbloc/hub-router v0.0.0-00010101000000-000000000000
	gotest.tools/v3 v3.0.3 // indirect
)

replace (
	github.com/trustbloc/hub-router => ../..
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
)
