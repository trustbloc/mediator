/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	didexdsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatordsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/trustbloc/edge-core/pkg/log"
	"github.com/trustbloc/edge-core/pkg/storage"

	"github.com/trustbloc/hub-router/pkg/aries"
	"github.com/trustbloc/hub-router/pkg/internal/common/support"
	"github.com/trustbloc/hub-router/pkg/restapi/internal/httputil"
)

// API endpoints.
const (
	healthCheckPath = "/healthcheck"
	invitationPath  = "/didcomm/invitation"
)

// Msg svc constants.
const (
	msgTypeBaseURI        = "https://trustbloc.github.io/blinded-routing/1.0"
	createConnReq         = msgTypeBaseURI + "/create-conn-req"
	createConnResp        = msgTypeBaseURI + "/create-conn-resp"
	createConnReqPurpose  = "create-conn-req"
	createConnRespPurpose = "create-conn-resp"
)

var logger = log.New("hub-router/operations")

// Handler http handler for each controller API endpoint.
type Handler interface {
	Path() string
	Method() string
	Handle() http.HandlerFunc
}

// Storage config.
type Storage struct {
	Persistent storage.Provider
	Transient  storage.Provider
}

// Config holds configuration.
type Config struct {
	Aries          aries.Ctx
	AriesMessenger service.Messenger
	MsgRegistrar   *msghandler.Registrar
	Storage        *Storage
}

// Operation implements hub-router operations.
type Operation struct {
	storage      *Storage
	oob          aries.OutOfBand
	didExchange  aries.DIDExchange
	mediator     aries.Mediator
	messenger    service.Messenger
	vdriRegistry vdri.Registry
	endpoint     string
}

// New returns a new Operation.
func New(config *Config) (*Operation, error) {
	actionCh := make(chan service.DIDCommAction, 1)

	oobClient, err := aries.CreateOutofbandClient(config.Aries)
	if err != nil {
		return nil, fmt.Errorf("out-of-band client: %w", err)
	}

	mediatorClient, err := aries.CreateMediatorClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("mediator client: %w", err)
	}

	didExchangeClient, err := aries.CreateDIDExchangeClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	o := &Operation{
		storage:      config.Storage,
		oob:          oobClient,
		didExchange:  didExchangeClient,
		mediator:     mediatorClient,
		messenger:    config.AriesMessenger,
		vdriRegistry: config.Aries.VDRIRegistry(),
		endpoint:     config.Aries.RouterEndpoint(),
	}

	msgCh := make(chan service.DIDCommMsg, 1)

	msgSvc := aries.NewMsgSvc("create-connection", createConnReq, createConnReqPurpose, msgCh)

	err = config.MsgRegistrar.Register(msgSvc)
	if err != nil {
		return nil, fmt.Errorf("message service client: %w", err)
	}

	go o.didCommActionListener(actionCh)

	go o.didCommMsgListener(msgCh)

	return o, nil
}

// GetRESTHandlers get all controller API handler available for this service.
func (o *Operation) GetRESTHandlers() []Handler {
	return []Handler{
		// healthcheck
		support.NewHTTPHandler(healthCheckPath, http.MethodGet, o.healthCheckHandler),

		// router
		support.NewHTTPHandler(invitationPath, http.MethodGet, o.generateInvitation),
	}
}

func (o *Operation) healthCheckHandler(rw http.ResponseWriter, _ *http.Request) {
	resp := &healthCheckResp{
		Status:      "success",
		CurrentTime: time.Now(),
	}

	httputil.WriteResponseWithLog(rw, resp, healthCheckPath, logger)
}

func (o *Operation) generateInvitation(rw http.ResponseWriter, _ *http.Request) {
	// TODO configure hub-router label
	invitation, err := o.oob.CreateInvitation(nil, outofband.WithLabel("hub-router"))
	if err != nil {
		httputil.WriteErrorResponseWithLog(rw, http.StatusInternalServerError,
			fmt.Sprintf("failed to create router invitation - err=%s", err.Error()), invitationPath, logger)

		return
	}

	httputil.WriteResponseWithLog(rw, &DIDCommInvitationResp{
		Invitation: invitation,
	}, invitationPath, logger)
}

func (o *Operation) didCommActionListener(ch <-chan service.DIDCommAction) {
	for msg := range ch {
		var err error

		var args interface{}

		switch msg.Message.Type() {
		case didexdsvc.RequestMsgType:
			args = nil
		case mediatordsvc.RequestMsgType:
			args = nil
		default:
			err = fmt.Errorf("unsupported message type : %s", msg.Message.Type())
		}

		if err != nil {
			logger.Errorf("msgType=[%s] id=[%s] errMsg=[%s]", msg.Message.Type(), msg.Message.ID(), err.Error())

			msg.Stop(fmt.Errorf("handle %s : %w", msg.Message.Type(), err))
		} else {
			logger.Infof("msgType=[%s] id=[%s] msg=[%s]", msg.Message.Type(), msg.Message.ID(), "success")

			msg.Continue(args)
		}
	}
}

func (o *Operation) didCommMsgListener(ch <-chan service.DIDCommMsg) {
	for msg := range ch {
		var err error

		var msgMap service.DIDCommMsgMap

		switch msg.Type() {
		case createConnReq:
			msgMap, err = o.handleCreateConnReq(msg)
		default:
			err = fmt.Errorf("unsupported message service type : %s", msg.Type())
		}

		if err != nil {
			msgMap = service.NewDIDCommMsgMap(&CreateConnResp{
				ID:      uuid.New().String(),
				Type:    createConnResp,
				Purpose: []string{createConnRespPurpose},
				Data:    &CreateConnRespData{ErrorMsg: err.Error()},
			})

			logger.Errorf("msgType=[%s] id=[%s] errMsg=[%s]", msg.Type(), msg.ID(), err.Error())
		}

		err = o.messenger.ReplyTo(msg.ID(), msgMap)
		if err != nil {
			logger.Errorf("sendReply : msgType=[%s] id=[%s] errMsg=[%s]", msg.Type(), msg.ID(), err.Error())

			continue
		}

		logger.Infof("msgType=[%s] id=[%s] msg=[%s]", msg.Type(), msg.ID(), "success")
	}
}

func (o *Operation) handleCreateConnReq(msg service.DIDCommMsg) (service.DIDCommMsgMap, error) {
	pMsg := CreateConnReq{}

	err := msg.Decode(&pMsg)
	if err != nil {
		return nil, fmt.Errorf("parse didcomm message : %w", err)
	}

	// get the peerDID from the request
	if pMsg.Data == nil || pMsg.Data.DIDDoc == nil {
		return nil, errors.New("did document mandatory")
	}

	docBytes, err := json.Marshal(pMsg.Data.DIDDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to read did document bytes : %w", err)
	}

	didDoc, err := did.ParseDocument(docBytes)
	if err != nil {
		return nil, fmt.Errorf("parse did doc : %w", err)
	}

	// create peer DID
	newDidDoc, err := o.vdriRegistry.Create("peer", vdri.WithServices(did.Service{ServiceEndpoint: o.endpoint}))
	if err != nil {
		return nil, fmt.Errorf("create new peer did : %w", err)
	}

	// create connection
	_, err = o.didExchange.CreateConnection(newDidDoc.ID, didDoc)
	if err != nil {
		return nil, fmt.Errorf("create connection : %w", err)
	}

	newDocBytes, err := newDidDoc.JSONBytes()
	if err != nil {
		return nil, fmt.Errorf("marshal did doc : %w", err)
	}

	// send router did doc
	return service.NewDIDCommMsgMap(&CreateConnResp{
		ID:      uuid.New().String(),
		Type:    createConnResp,
		Purpose: []string{createConnRespPurpose},
		Data:    &CreateConnRespData{DIDDoc: newDocBytes},
	}), nil
}
