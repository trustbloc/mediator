/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package outofbandv2

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/outofbandv2"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
)

// MockClient is a mock out-of-band 2.0 client used in tests.
type MockClient struct {
	CreateValue *outofbandsvc.Invitation
	CreateErr   error
	AcceptValue string
	AcceptErr   error
}

// CreateInvitation creates a mock outofbandv2 invitation.
func (c *MockClient) CreateInvitation(...outofbandv2.MessageOption) (*outofbandsvc.Invitation, error) {
	return c.CreateValue, c.CreateErr
}

// AcceptInvitation accepts invitation.
func (c *MockClient) AcceptInvitation(_ *outofbandsvc.Invitation) (string, error) {
	return c.AcceptValue, c.AcceptErr
}
