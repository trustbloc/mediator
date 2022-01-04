/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package outofbandv2

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
)

// MockService is a mock outofband 2.0 service.
type MockService struct {
	AcceptInvitationValue string
	AcceptInvitationErr   error
	SaveInvitationErr     error
}

// AcceptInvitation accepts invitation.
func (m *MockService) AcceptInvitation(*outofbandv2.Invitation) (string, error) {
	return m.AcceptInvitationValue, m.AcceptInvitationErr
}

// SaveInvitation saves invitation.
func (m *MockService) SaveInvitation(*outofbandv2.Invitation) error {
	return m.SaveInvitationErr
}
