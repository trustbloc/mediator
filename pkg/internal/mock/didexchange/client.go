/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didexchange

import (
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
)

// MockClient is a mock didexchange.MockClient used in tests.
type MockClient struct {
	ActionEventFunc  func(chan<- service.DIDCommAction) error
	CreateConnErr    error
	GetConnectionErr error
}

// RegisterActionEvent registers the action event channel.
func (c *MockClient) RegisterActionEvent(actions chan<- service.DIDCommAction) error {
	if c.ActionEventFunc != nil {
		return c.ActionEventFunc(actions)
	}

	return nil
}

// CreateConnection creates connection.
func (c *MockClient) CreateConnection(_ string, _ *did.Doc, _ ...didexchange.ConnectionOption) (string, error) {
	if c.CreateConnErr != nil {
		return "", c.CreateConnErr
	}

	return uuid.New().String(), nil
}

// GetConnection fetches connection record based on connID.
func (c *MockClient) GetConnection(connectionID string) (*didexchange.Connection, error) {
	if c.GetConnectionErr != nil {
		return nil, c.GetConnectionErr
	}

	return &didexchange.Connection{Record: &connection.Record{ConnectionID: connectionID}}, nil
}
