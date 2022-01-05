/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"crypto/tls"

	tlsutils "github.com/trustbloc/edge-core/pkg/utils/tls"
)

// BDDContext is a global context shared between different test suites in bddtests.
type BDDContext struct {
	TLSConfig   *tls.Config
	SidetreeURL string
}

// NewBDDContext create new BDDContext.
func NewBDDContext(caCertPath string) (*BDDContext, error) {
	rootCAs, err := tlsutils.GetCertPool(false, []string{caCertPath})
	if err != nil {
		return nil, err
	}

	return &BDDContext{
		TLSConfig:   &tls.Config{RootCAs: rootCAs, MinVersion: tls.VersionTLS12},
		SidetreeURL: "http://localhost:48326/sidetree/v1/",
	}, nil
}
