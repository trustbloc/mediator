/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import "github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"

// MsgService msg service implementation.
type MsgService struct {
	svcName string
	purpose []string
	msgType string
	msgCh   chan service.DIDCommMsg
}

// NewMsgSvc new msg service.
func NewMsgSvc(name, msgType, purpose string, msgCh chan service.DIDCommMsg) *MsgService {
	return &MsgService{
		svcName: name,
		msgType: msgType,
		purpose: []string{purpose},
		msgCh:   msgCh,
	}
}

// Name svc name.
func (m *MsgService) Name() string {
	return m.svcName
}

// Accept validates whether the service handles msgType and purpose.
func (m *MsgService) Accept(msgType string, purpose []string) bool {
	purposeMatched, typeMatched := len(m.purpose) == 0, m.msgType == ""

	if purposeMatched && typeMatched {
		return false
	}

	for _, purposeCriteria := range m.purpose {
		for _, msgPurpose := range purpose {
			if purposeCriteria == msgPurpose {
				purposeMatched = true

				break
			}
		}
	}

	if m.msgType == msgType {
		typeMatched = true
	}

	return purposeMatched && typeMatched
}

// HandleInbound handles inbound didcomm msg.
func (m *MsgService) HandleInbound(msg service.DIDCommMsg, myDID, theirDID string) (string, error) {
	go func() {
		m.msgCh <- msg
	}()

	return "", nil
}
