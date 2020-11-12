/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
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
	t.Run("success", func(t *testing.T) {
		c, err := CreateDIDExchangeClient(getAriesCtx(), nil, nil)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("client creation error", func(t *testing.T) {
		c, err := CreateDIDExchangeClient(&mockprovider.Provider{}, nil, nil)
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create didexchange client")
	})

	t.Run("action event registration error", func(t *testing.T) {
		ctx := &mockprovider.Provider{
			ProtocolStateStorageProviderValue: mockstore.NewMockStoreProvider(),
			StorageProviderValue:              mockstore.NewMockStoreProvider(),
			ServiceMap: map[string]interface{}{
				didexchange.DIDExchange: &mocksvc.MockDIDExchangeSvc{RegisterActionEventErr: errors.New("reg error")},
				mediator.Coordination:   &mockroute.MockMediatorSvc{},
			},
		}

		c, err := CreateDIDExchangeClient(ctx, make(chan service.DIDCommAction), nil)
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "register didexchange action event")
	})

	t.Run("msg event registration error", func(t *testing.T) {
		ctx := &mockprovider.Provider{
			ProtocolStateStorageProviderValue: mockstore.NewMockStoreProvider(),
			StorageProviderValue:              mockstore.NewMockStoreProvider(),
			ServiceMap: map[string]interface{}{
				didexchange.DIDExchange: &mocksvc.MockDIDExchangeSvc{RegisterMsgEventErr: errors.New("reg msg error")},
				mediator.Coordination:   &mockroute.MockMediatorSvc{},
			},
		}

		c, err := CreateDIDExchangeClient(ctx, make(chan service.DIDCommAction), nil)
		require.Nil(t, c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "register didexchange message event")
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
