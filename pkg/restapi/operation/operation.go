/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
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
	didExchange aries.DIDExchange
}

// New returns a new Operation.
func New(config *Config) (*Operation, error) {
	actionCh := make(chan service.DIDCommAction, 1)

	didExchangeClient, err := aries.CreateDIDExchangeClient(config.Aries, actionCh)
	if err != nil {
		return nil, fmt.Errorf("didexchange client: %w", err)
	}

	o := &Operation{
		storage:     config.Storage,
		didExchange: didExchangeClient,
	}

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
	invitation, err := o.didExchange.CreateInvitation("hub-router")
	if err != nil {
		httputil.WriteErrorResponseWithLog(rw, http.StatusInternalServerError,
			fmt.Sprintf("failed to create router invitation - err=%s", err.Error()), invitationPath, logger)

		return
	}

	httputil.WriteResponseWithLog(rw, invitation, invitationPath, logger)
}
