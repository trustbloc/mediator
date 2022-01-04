/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexdsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatordsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockdiddoc "github.com/hyperledger/aries-framework-go/pkg/mock/diddoc"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	mockvdri "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/hub-router/pkg/internal/mock/didexchange"
	"github.com/trustbloc/hub-router/pkg/internal/mock/messenger"
	mockoutofband "github.com/trustbloc/hub-router/pkg/internal/mock/outofband"
	mockoutofbandv2 "github.com/trustbloc/hub-router/pkg/internal/mock/outofbandv2"
)

func TestNew(t *testing.T) {
	t.Run("returns instance", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		require.Len(t, o.GetRESTHandlers(), 3)
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

	t.Run("out of band v2 client creation error", func(t *testing.T) {
		config := config()
		config.Aries = &mockprovider.Provider{
			ServiceMap: map[string]interface{}{
				outofbandsvc.Name:         &mockoutofband.MockService{},
				mediatordsvc.Coordination: &mockroute.MockMediatorSvc{},
				didexdsvc.DIDExchange:     &mocksvc.MockDIDExchangeSvc{},
			},
		}

		o, err := New(config)
		require.Nil(t, o)
		require.Error(t, err)
		require.Contains(t, err.Error(), "out-of-band-v2 client")
	})

	t.Run("mediator client creation error", func(t *testing.T) {
		config := config()
		config.Aries = &mockprovider.Provider{
			ServiceMap: map[string]interface{}{
				outofbandsvc.Name:     &mockoutofband.MockService{},
				outofbandv2.Name:      &mockoutofbandv2.MockService{},
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
				outofbandv2.Name:          &mockoutofbandv2.MockService{},
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
		require.Equal(t, result.Invitation.Type, "https://didcomm.org/out-of-band/1.0/invitation")
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

func TestGenerateInvitationV2Handler(t *testing.T) {
	reqData := DIDCommInvitationV2Req{
		DID: "did:foo:bar",
	}

	reqBytes, err := json.Marshal(&reqData)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, invitationV2Path, bytes.NewReader(reqBytes))

		w := httptest.NewRecorder()
		o.generateInvitationV2(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var result *DIDCommInvitationV2Resp
		err = json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		require.NotEmpty(t, result.Invitation.ID)
		require.Equal(t, result.Invitation.Label, "hub-router")
		require.Equal(t, result.Invitation.Type, "https://didcomm.org/out-of-band/2.0/invitation")
	})

	t.Run("parse error", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, invitationV2Path, bytes.NewReader([]byte("bad data")))

		w := httptest.NewRecorder()
		o.generateInvitationV2(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "error parsing request")
	})

	t.Run("create error", func(t *testing.T) {
		o, err := New(config())
		require.NoError(t, err)

		o.oobv2 = &mockoutofbandv2.MockClient{CreateErr: errors.New("invitation error")}

		req := httptest.NewRequest(http.MethodPost, invitationV2Path, bytes.NewReader(reqBytes))

		w := httptest.NewRecorder()
		o.generateInvitationV2(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "error creating invitation")
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

func TestDIDCommStateMsgListener(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		done := make(chan struct{})

		c.messenger = &messenger.MockMessenger{
			SendFunc: func(msg service.DIDCommMsgMap, myDID, theirDID string) error {
				pMsg := &DIDCommMsg{}
				err = msg.Decode(pMsg)
				require.NoError(t, err)

				done <- struct{}{}

				return nil
			},
		}
		c.didExchange = &didexchange.MockClient{}

		msgCh := make(chan service.StateMsg, 1)
		go c.stateMsgHandler(msgCh)

		msgCh <- service.StateMsg{
			Type:         service.PostState,
			ProtocolName: didexdsvc.DIDExchange,
			StateID:      didexdsvc.StateIDCompleted,
			Properties: &didexchangeEvent{
				connID: uuid.New().String(),
			},
		}

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})

	t.Run("ignore pre state", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		msg := service.StateMsg{
			Type:         service.PreState,
			ProtocolName: didexdsvc.DIDExchange,
			StateID:      didexdsvc.StateIDCompleted,
			Properties: &didexchangeEvent{
				connID: uuid.New().String(),
			},
		}

		err = c.hanlDIDExStateMsg(msg)
		require.NoError(t, err)
	})

	t.Run("send message error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.messenger = &messenger.MockMessenger{
			SendFunc: func(msg service.DIDCommMsgMap, myDID, theirDID string) error {
				return errors.New("send error")
			},
		}
		c.didExchange = &didexchange.MockClient{}

		msg := service.StateMsg{
			Type:         service.PostState,
			ProtocolName: didexdsvc.DIDExchange,
			StateID:      didexdsvc.StateIDCompleted,
			Properties: &didexchangeEvent{
				connID: uuid.New().String(),
			},
		}

		err = c.hanlDIDExStateMsg(msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "send didex state complete msg")
	})

	t.Run("cast to didex event error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.messenger = &messenger.MockMessenger{
			SendFunc: func(msg service.DIDCommMsgMap, myDID, theirDID string) error {
				return errors.New("send error")
			},
		}
		c.didExchange = &didexchange.MockClient{}

		msg := service.StateMsg{
			Type:         service.PostState,
			ProtocolName: didexdsvc.DIDExchange,
			StateID:      didexdsvc.StateIDCompleted,
		}

		err = c.hanlDIDExStateMsg(msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to cast didexchange event properties")
	})

	t.Run("get connection error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.messenger = &messenger.MockMessenger{
			SendFunc: func(msg service.DIDCommMsgMap, myDID, theirDID string) error {
				return errors.New("send error")
			},
		}
		c.didExchange = &didexchange.MockClient{
			GetConnectionErr: errors.New("get conn error"),
		}

		msg := service.StateMsg{
			Type:         service.PostState,
			ProtocolName: didexdsvc.DIDExchange,
			StateID:      didexdsvc.StateIDCompleted,
			Properties: &didexchangeEvent{
				connID: uuid.New().String(),
			},
		}

		err = c.hanlDIDExStateMsg(msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get connection for id=")
	})
}

func TestDIDCommMsgListener(t *testing.T) {
	t.Run("unsupported message type", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		done := make(chan struct{})

		c.messenger = &messenger.MockMessenger{
			ReplyToFunc: func(msgID string, msg service.DIDCommMsgMap, _ ...service.Opt) error {
				pMsg := &CreateConnResp{}
				err = msg.Decode(pMsg)
				require.NoError(t, err)

				require.Contains(t, pMsg.Data.ErrorMsg, "unsupported message service type : unsupported-message-type")
				require.Empty(t, pMsg.Data.DIDDoc)

				done <- struct{}{}

				return nil
			},
		}

		msgCh := make(chan service.DIDCommMsg, 1)
		go c.didCommMsgListener(msgCh)

		msgCh <- service.NewDIDCommMsgMap(struct {
			Type string `json:"@type,omitempty"`
		}{Type: "unsupported-message-type"})

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})

	t.Run("messenger reply error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.messenger = &messenger.MockMessenger{
			ReplyToFunc: func(msgID string, msg service.DIDCommMsgMap, _ ...service.Opt) error {
				return errors.New("reply error")
			},
		}

		msgCh := make(chan service.DIDCommMsg, 1)
		go c.didCommMsgListener(msgCh)

		msgCh <- service.NewDIDCommMsgMap(struct {
			Type string `json:"@type,omitempty"`
		}{Type: "unsupported-message-type"})
	})

	t.Run("create connection request", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		done := make(chan struct{})

		c.messenger = &messenger.MockMessenger{
			ReplyToFunc: func(msgID string, msg service.DIDCommMsgMap, _ ...service.Opt) error {
				pMsg := &CreateConnResp{}
				dErr := msg.Decode(pMsg)
				require.NoError(t, dErr)

				docBytes, dErr := json.Marshal(pMsg.Data.DIDDoc)
				require.NoError(t, dErr)

				didDoc, dErr := did.ParseDocument(docBytes)
				require.NoError(t, dErr)

				require.Contains(t, didDoc.ID, "did:")
				require.Equal(t, pMsg.Type, createConnResp)

				done <- struct{}{}

				return nil
			},
		}

		msgCh := make(chan service.DIDCommMsg, 1)
		go c.didCommMsgListener(msgCh)

		didDocBytes, err := mockdiddoc.GetMockDIDDoc(t).JSONBytes()
		require.NoError(t, err)

		msgCh <- service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{
				DIDDoc: json.RawMessage(didDocBytes),
			},
		})

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "tests are not validated due to timeout")
		}
	})
}

func TestCreateConnectionReqHanlder(t *testing.T) {
	t.Run("no did doc", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		msg := service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{},
		})

		_, err = c.handleCreateConnReq(msg)
		require.Contains(t, err.Error(), "did document mandatory")
	})

	t.Run("invalid did doc error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		msg := service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{
				DIDDoc: []byte("invalid-diddoc"),
			},
		})

		_, err = c.handleCreateConnReq(msg)
		require.Contains(t, err.Error(), "parse did doc")
	})

	t.Run("invalid did doc", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.vdriRegistry = &mockvdri.MockVDRegistry{
			CreateErr: errors.New("did create error"),
		}

		didDocBytes, err := mockdiddoc.GetMockDIDDoc(t).JSONBytes()
		require.NoError(t, err)

		msg := service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{
				DIDDoc: json.RawMessage(didDocBytes),
			},
		})

		_, err = c.handleCreateConnReq(msg)
		require.Contains(t, err.Error(), "create new peer did")
	})

	t.Run("create conn error", func(t *testing.T) {
		c, err := New(config())
		require.NoError(t, err)

		c.didExchange = &didexchange.MockClient{
			CreateConnErr: errors.New("create error"),
		}

		didDocBytes, err := mockdiddoc.GetMockDIDDoc(t).JSONBytes()
		require.NoError(t, err)

		msg := service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{
				DIDDoc: json.RawMessage(didDocBytes),
			},
		})

		_, err = c.handleCreateConnReq(msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create connection")
	})

	t.Run("error if cannot create key", func(t *testing.T) {
		expected := errors.New("test")

		c, err := New(config())
		require.NoError(t, err)

		c.keyManager = &mockkms.KeyManager{CrAndExportPubKeyErr: expected}

		didDocBytes, err := mockdiddoc.GetMockDIDDoc(t).JSONBytes()
		require.NoError(t, err)

		msg := service.NewDIDCommMsgMap(CreateConnReq{
			ID:   uuid.New().String(),
			Type: createConnReq,
			Data: &CreateConnReqData{
				DIDDoc: didDocBytes,
			},
		})

		_, err = c.handleCreateConnReq(msg)
		require.ErrorIs(t, err, expected)
	})
}
