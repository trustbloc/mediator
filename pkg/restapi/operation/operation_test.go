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
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexdsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatordsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"

	mockoutofband "github.com/trustbloc/hub-router/pkg/internal/mock/outofband"
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
		require.Contains(t, err.Error(), "out-of-band client")
	})

	t.Run("mediator client creation error", func(t *testing.T) {
		config := config()
		config.Aries = &mockprovider.Provider{
			ServiceMap: map[string]interface{}{
				outofbandsvc.Name:     &mockoutofband.MockService{},
				didexdsvc.DIDExchange: &mocksvc.MockDIDExchangeSvc{},
			},
		}

		o, err := New(config)
		require.Nil(t, o)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mediator client")
	})

	t.Run("didex client creation error", func(t *testing.T) {
		config := config()
		config.Aries = &mockprovider.Provider{
			ServiceMap: map[string]interface{}{
				outofbandsvc.Name:         &mockoutofband.MockService{},
				mediatordsvc.Coordination: &mockroute.MockMediatorSvc{},
			},
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

		var result *DIDCommInvitationResp
		err = json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		require.NotEmpty(t, result.Invitation.ID)
		require.Equal(t, result.Invitation.Label, "hub-router")
		require.Equal(t, result.Invitation.Type, "https://didcomm.org/oob-invitation/1.0/invitation")
	})

	t.Run("error", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		o.oob = &mockoutofband.MockClient{CreateInvitationErr: errors.New("invitation error")}

		w := httptest.NewRecorder()
		o.generateInvitation(w, nil)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "failed to create router invitation")
	})
}

func TestDIDCommListener(t *testing.T) {
	c, err := New(config())
	require.NoError(t, err)

	actionCh := make(chan service.DIDCommAction, 1)
	go c.didCommActionListener(actionCh)

	t.Run("didexchange request", func(t *testing.T) {
		done := make(chan struct{})

		actionCh <- service.DIDCommAction{
			Message: service.NewDIDCommMsgMap(struct {
				Type string `json:"@type,omitempty"`
			}{Type: didexdsvc.RequestMsgType}),
			Continue: func(args interface{}) {
				require.Nil(t, args)

				done <- struct{}{}
			},
		}

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})

	t.Run("mediation request", func(t *testing.T) {
		done := make(chan struct{})

		actionCh <- service.DIDCommAction{
			Message: service.NewDIDCommMsgMap(struct {
				Type string `json:"@type,omitempty"`
			}{Type: mediatordsvc.RequestMsgType}),
			Continue: func(args interface{}) {
				require.Nil(t, args)

				done <- struct{}{}
			},
		}

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})

	t.Run("unsupported message type", func(t *testing.T) {
		done := make(chan struct{})

		actionCh <- service.DIDCommAction{
			Message: service.NewDIDCommMsgMap(struct {
				Type string `json:"@type,omitempty"`
			}{Type: "unsupported-message-type"}),
			Stop: func(err error) {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), "unsupported message type")
				done <- struct{}{}
			},
		}

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})
}
