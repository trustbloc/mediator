/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"

	mockdidex "github.com/trustbloc/hub-router/pkg/internal/mock/didexchange"
)

func TestNew(t *testing.T) {
	t.Run("returns instance", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		require.Len(t, o.GetRESTHandlers(), 2)
	})

	t.Run("aries store error", func(t *testing.T) {
		config := config()
		config.Aries = &mockprovider.Provider{
			StorageProviderValue: mockstore.NewMockStoreProvider(),
		}

		o, err := New(config)
		require.Nil(t, o)
		require.Error(t, err)
		require.Contains(t, err.Error(), "didexchange client")
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

func TestGenerateInvitationHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		w := httptest.NewRecorder()
		o.generateInvitation(w, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var invitation didexchange.Invitation
		err = json.Unmarshal(w.Body.Bytes(), &invitation)
		require.NoError(t, err)

		require.NotEmpty(t, invitation.ID)
		require.Equal(t, invitation.Label, "hub-router")
		require.Equal(t, invitation.Type, "https://didcomm.org/didexchange/1.0/invitation")
	})

	t.Run("error", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		o.didExchange = &mockdidex.MockClient{CreateInvitationErr: errors.New("invitation error")}

		w := httptest.NewRecorder()
		o.generateInvitation(w, nil)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "failed to create router invitation")
	})
}
