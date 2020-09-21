/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didexchange

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
)

// MockClient is a mock didexchange.MockClient used in tests.
type MockClient struct {
	ActionEventFunc     func(chan<- service.DIDCommAction) error
	CreateInvitationErr error
}

// RegisterActionEvent registers the action event channel.
func (c *MockClient) RegisterActionEvent(actions chan<- service.DIDCommAction) error {
	if c.ActionEventFunc != nil {
		return c.ActionEventFunc(actions)
	}

	return nil
}

// CreateInvitation creates an explicit invitation.
func (c *MockClient) CreateInvitation(label string) (*didexchange.Invitation, error) {
	if c.CreateInvitationErr != nil {
		return nil, c.CreateInvitationErr
	}

	return &didexchange.Invitation{}, nil
}
