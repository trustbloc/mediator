/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package outofband

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
)

// MockService is a mock outofband service.
type MockService struct {
	SaveInvitationErr error
}

// RegisterActionEvent mock.
func (m *MockService) RegisterActionEvent(ch chan<- service.DIDCommAction) error {
	panic("not implemented")
}

// UnregisterActionEvent mock.
func (m *MockService) UnregisterActionEvent(ch chan<- service.DIDCommAction) error {
	panic("not implemented")
}

// RegisterMsgEvent mock.
func (m *MockService) RegisterMsgEvent(ch chan<- service.StateMsg) error {
	panic("not implemented")
}

// UnregisterMsgEvent mock.
func (m *MockService) UnregisterMsgEvent(ch chan<- service.StateMsg) error {
	panic("not implemented")
}

// AcceptInvitation mock.
func (m *MockService) AcceptInvitation(invitation *outofband.Invitation, s string,
	routerConnections []string) (string, error) {
	panic("not implemented")
}

// SaveInvitation mock.
func (m *MockService) SaveInvitation(invitation *outofband.Invitation) error {
	return m.SaveInvitationErr
}

// Actions mock.
func (m *MockService) Actions() ([]outofband.Action, error) {
	panic("not implemented")
}

// ActionContinue mock.
func (m *MockService) ActionContinue(s string, options outofband.Options) error {
	panic("not implemented")
}

// ActionStop mock.
func (m *MockService) ActionStop(s string, err error) error {
	panic("not implemented")
}
