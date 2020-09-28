/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	"github.com/stretchr/testify/require"
)

func TestCreateOutofbandClient(t *testing.T) {
	t.Run("oob client - success", func(t *testing.T) {
		c, err := CreateOutofbandClient(getAriesCtx())
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("oob client - error", func(t *testing.T) {
		c, err := CreateOutofbandClient(&mockprovider.Provider{})
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create out-of-band client")
	})
}

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

func TestCreateMediatorClient(t *testing.T) {
	actionCh := make(chan service.DIDCommAction)

	t.Run("mediator client - success", func(t *testing.T) {
		c, err := CreateMediatorClient(getAriesCtx(), actionCh)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("mediator client - error", func(t *testing.T) {
		c, err := CreateMediatorClient(&mockprovider.Provider{}, actionCh)
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create mediator client")
	})
}
