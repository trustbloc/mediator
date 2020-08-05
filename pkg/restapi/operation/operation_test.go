/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"
	"github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	ariesmockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/edge-core/pkg/storage/memstore"
)

func TestNew(t *testing.T) {
	t.Run("returns instance", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		require.Len(t, o.GetRESTHandlers(), 1)
	})

	t.Run("aries store error", func(t *testing.T) {
		config := config()
		config.Aries = &protocol.MockProvider{
			StoreProvider: &ariesmockstore.MockStoreProvider{
				ErrOpenStoreHandle: errors.New("test"),
			},
		}

		_, err := New(config)
		require.Error(t, err)
	})
}

func TestOperation_HealthCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		w := httptest.NewRecorder()
		o.healthCheckHandler(w, nil)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

func config() *Config {
	return &Config{
		Aries: &protocol.MockProvider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination: &mediator.MockMediatorSvc{},
			},
		},
		Storage: &Storage{
			Persistent: memstore.NewProvider(),
			Transient:  memstore.NewProvider(),
		},
	}
}
