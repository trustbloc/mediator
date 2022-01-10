/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofbandv2"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	didexdsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatordsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/peer"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/hub-router/pkg/aries"
	"github.com/trustbloc/hub-router/pkg/internal/common/support"
	"github.com/trustbloc/hub-router/pkg/restapi/internal/httputil"
)

// API endpoints.
const (
	healthCheckPath  = "/healthcheck"
	invitationPath   = "/didcomm/invitation"
	invitationV2Path = "/didcomm/invitation-v2"
)

// Msg svc constants.
const (
	msgTypeBaseURI    = "https://trustbloc.dev"
	blindedRoutingURI = msgTypeBaseURI + "/blinded-routing/1.0"
	createConnReq     = blindedRoutingURI + "/create-conn-req"
	createConnResp    = blindedRoutingURI + "/create-conn-resp"
	didExStateComp    = msgTypeBaseURI + "/didexchange/1.0/state-complete"
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
	PublicDID      string
}

// Operation implements hub-router operations.
type Operation struct {
	storage      *Storage
	oob          aries.OutOfBand
	oobv2        aries.OutOfBandV2
	didExchange  aries.DIDExchange
	mediator     aries.Mediator
	messenger    service.Messenger
	vdriRegistry vdrapi.Registry
	keyManager   kms.KeyManager
	endpoint     string
	publicDID    string
}

// New returns a new Operation.
func New(config *Config) (*Operation, error) {
	actionCh := make(chan service.DIDCommAction, 1)
	stateMsgCh := make(chan service.StateMsg, 1)

	oobClient, err := aries.CreateOutofbandClient(config.Aries)
	if err != nil {
		return nil, fmt.Errorf("out-of-band client: %w", err)
	}

	oobV2Client, err := aries.CreateOutOfBandV2Client(config.Aries)
	if err != nil {
		return nil, fmt.Errorf("out-of-band-v2 client: %w", err)
	}

	mediatorClient, err := aries.CreateMediatorClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("mediator client: %w", err)
	}

	didExchangeClient, err := aries.CreateDIDExchangeClient(config.Aries, actionCh, stateMsgCh)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	o := &Operation{
		storage:      config.Storage,
		oob:          oobClient,
		oobv2:        oobV2Client,
		didExchange:  didExchangeClient,
		mediator:     mediatorClient,
		messenger:    config.AriesMessenger,
		vdriRegistry: config.Aries.VDRegistry(),
		endpoint:     config.Aries.RouterEndpoint(),
		keyManager:   config.Aries.KMS(),
		publicDID:    config.PublicDID,
	}

	msgCh := make(chan service.DIDCommMsg, 1)

	msgSvc := aries.NewMsgSvc("create-connection", createConnReq, msgCh)

	err = config.MsgRegistrar.Register(msgSvc)
	if err != nil {
		return nil, fmt.Errorf("message service client: %w", err)
	}

	go o.didCommActionListener(actionCh)

	go o.didCommMsgListener(msgCh)

	go o.stateMsgHandler(stateMsgCh)

	return o, nil
}

// GetRESTHandlers get all controller API handler available for this service.
func (o *Operation) GetRESTHandlers() []Handler {
	return []Handler{
		// healthcheck
		support.NewHTTPHandler(healthCheckPath, http.MethodGet, o.healthCheckHandler),

		// router
		support.NewHTTPHandler(invitationPath, http.MethodGet, o.generateInvitation),
		support.NewHTTPHandler(invitationV2Path, http.MethodGet, o.generateInvitationV2),
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

func (o *Operation) generateInvitationV2(rw http.ResponseWriter, _ *http.Request) {
	// TODO configure hub-router label
	invitation, err := o.oobv2.CreateInvitation(
		outofbandv2.WithFrom(o.publicDID),
		outofbandv2.WithLabel("hub-router"),
		outofbandv2.WithAccept(transport.MediaTypeDIDCommV2Profile, transport.MediaTypeAIP2RFC0019Profile),
	)
	if err != nil {
		httputil.WriteErrorResponseWithLog(rw, http.StatusInternalServerError,
			"error creating invitation", invitationV2Path, logger)

		return
	}

	httputil.WriteResponseWithLog(rw, &DIDCommInvitationV2Resp{
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
				ID:   uuid.New().String(),
				Type: createConnResp,
				Data: &CreateConnRespData{ErrorMsg: err.Error()},
			})

			logger.Errorf("msgType=[%s] id=[%s] errMsg=[%s]", msg.Type(), msg.ID(), err.Error())
		}

		err = o.messenger.ReplyTo(msg.ID(), msgMap) // nolint:staticcheck //issue#47
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
	if pMsg.Data == nil || pMsg.Data.DIDDoc == nil || len(pMsg.Data.DIDDoc) == 0 {
		return nil, errors.New("did document mandatory")
	}

	didDoc, err := did.ParseDocument(pMsg.Data.DIDDoc)
	if err != nil {
		return nil, fmt.Errorf("parse did doc : %w", err)
	}

	// TODO - key type should be configurable
	keyID, pubKeyBytes, err := o.keyManager.CreateAndExportPubKeyBytes(kms.ED25519Type)
	if err != nil {
		return nil, fmt.Errorf("kms failed to create key: %w", err)
	}

	// create peer DID
	docResolution, err := o.vdriRegistry.Create(
		peer.DIDMethod,
		&did.Doc{
			Service: []did.Service{{ServiceEndpoint: o.endpoint}},
			VerificationMethod: []did.VerificationMethod{*did.NewVerificationMethodFromBytes(
				"#"+keyID,
				"Ed25519VerificationKey2018",
				"",
				pubKeyBytes,
			)},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create new peer did : %w", err)
	}

	// create connection
	_, err = o.didExchange.CreateConnection(docResolution.DIDDocument.ID, didDoc)
	if err != nil {
		return nil, fmt.Errorf("create connection : %w", err)
	}

	newDocBytes, err := docResolution.DIDDocument.JSONBytes()
	if err != nil {
		return nil, fmt.Errorf("marshal did doc : %w", err)
	}

	logger.Debugf("created PEER DID: %s", newDocBytes)

	// send router did doc
	return service.NewDIDCommMsgMap(&CreateConnResp{
		ID:   uuid.New().String(),
		Type: createConnResp,
		Data: &CreateConnRespData{DIDDoc: newDocBytes},
	}), nil
}

func (o *Operation) stateMsgHandler(stateMsgCh chan service.StateMsg) {
	for msg := range stateMsgCh {
		switch msg.ProtocolName {
		case didexdsvc.DIDExchange:
			err := o.hanlDIDExStateMsg(msg)
			if err != nil {
				logger.Errorf("failed to handle did exchange state message : %s", err.Error())
			}
		default:
			logger.Warnf("failed to cast didexchange event properties")
		}
	}
}

func (o *Operation) hanlDIDExStateMsg(msg service.StateMsg) error {
	if msg.Type != service.PostState || msg.StateID != didexdsvc.StateIDCompleted {
		logger.Debugf("handle did exchange state msg : stateMsgType=%s stateID=%s",
			service.PostState, didexdsvc.StateIDCompleted)

		return nil
	}

	var event didexchange.Event

	switch p := msg.Properties.(type) {
	case didexchange.Event:
		event = p
	default:
		return errors.New("failed to cast didexchange event properties")
	}

	conn, err := o.didExchange.GetConnection(event.ConnectionID())
	if err != nil {
		return fmt.Errorf("get connection for id=%s : %w", event.ConnectionID(), err)
	}

	err = o.messenger.Send(service.NewDIDCommMsgMap(&DIDCommMsg{
		ID:   uuid.New().String(),
		Type: didExStateComp,
	}), conn.MyDID, conn.TheirDID)
	if err != nil {
		return fmt.Errorf("send didex state complete msg : %w", err)
	}

	return nil
}
