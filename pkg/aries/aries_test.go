/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"testing"

	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	"github.com/stretchr/testify/require"
)

func TestCreateDIDExchangeClient(t *testing.T) {
	t.Run("did exchange client - success", func(t *testing.T) {
		c, err := CreateDIDExchangeClient(getAriesCtx(), nil)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("did exchange client - error", func(t *testing.T) {
		c, err := CreateDIDExchangeClient(&mockprovider.Provider{}, nil)
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create didexchange client")
	})
}
