/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import "github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"

// MsgService msg service implementation.
type MsgService struct {
	svcName string
	msgType string
	msgCh   chan service.DIDCommMsg
}

// NewMsgSvc new msg service.
func NewMsgSvc(name, msgType string, msgCh chan service.DIDCommMsg) *MsgService {
	return &MsgService{
		svcName: name,
		msgType: msgType,
		msgCh:   msgCh,
	}
}

// Name svc name.
func (m *MsgService) Name() string {
	return m.svcName
}

// Accept validates whether the service handles msgType and purpose.
func (m *MsgService) Accept(msgType string, _ []string) bool {
	return m.msgType == msgType
}

// HandleInbound handles inbound didcomm msg.
func (m *MsgService) HandleInbound(msg service.DIDCommMsg, _, _ string) (string, error) {
	go func() {
		m.msgCh <- msg
	}()

	return "", nil
}
