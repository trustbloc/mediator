/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexdsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatordsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
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
	Aries   aries.Ctx
	Storage *Storage
}

// Operation implements hub-router operations.
type Operation struct {
	storage     *Storage
	oob         aries.OutOfBand
	didExchange aries.DIDExchange
	mediator    aries.Mediator
}

// New returns a new Operation.
func New(config *Config) (*Operation, error) {
	actionCh := make(chan service.DIDCommAction, 1)

	oobClient, err := aries.CreateOutofbandClient(config.Aries)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	didExchangeClient, err := aries.CreateDIDExchangeClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	mediatorClient, err := aries.CreateMediatorClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	o := &Operation{
		storage:     config.Storage,
		oob:         oobClient,
		didExchange: didExchangeClient,
		mediator:    mediatorClient,
	}

	go o.didCommActionListener(actionCh)

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
