/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/trustbloc/edge-core/pkg/log"
	"github.com/trustbloc/edge-core/pkg/storage"

	"github.com/trustbloc/hub-router/pkg/internal/common/support"
)

// API endpoints.
const (
	healthCheckPath = "/healthcheck"
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

// Aries framework context provider.
type Aries interface {
	Service(string) (interface{}, error)
}

// Config holds configuration.
type Config struct {
	Aries   Aries
	Storage *Storage
}

// Mediator client.
type Mediator interface {
	RegisterActionEvent(chan<- service.DIDCommAction) error
}

// Operation implements hub-router operations.
type Operation struct {
	storage  *Storage
	mediator Mediator
}

// New returns a new Operation.
func New(config *Config) (*Operation, error) {
	o := &Operation{
		storage: config.Storage,
	}

	var err error

	o.mediator, err = mediatorClient(config.Aries)
	if err != nil {
		return nil, fmt.Errorf("failed to init operations: %w", err)
	}

	actions := make(chan<- service.DIDCommAction)

	err = o.mediator.RegisterActionEvent(actions)
	if err != nil {
		return nil, fmt.Errorf("failed to init operations: failed to register for mediator actions events: %w", err)
	}

	return o, nil
}

// GetRESTHandlers get all controller API handler available for this service.
func (o *Operation) GetRESTHandlers() []Handler {
	return []Handler{
		support.NewHTTPHandler(healthCheckPath, http.MethodGet, o.healthCheckHandler),
	}
}

func (o *Operation) healthCheckHandler(rw http.ResponseWriter, _ *http.Request) {
	logger.Debugf("serving healthcheck request")

	rw.WriteHeader(http.StatusOK)

	err := json.NewEncoder(rw).Encode(&healthCheckResp{
		Status:      "success",
		CurrentTime: time.Now(),
	})
	if err != nil {
		logger.Errorf("healthcheck response failure, %s", err)
	}

	logger.Debugf("done servicing healthcheck request")
}

func mediatorClient(ctx Aries) (Mediator, error) {
	c, err := mediator.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init aries mediator client: %w", err)
	}

	return c, nil
}
