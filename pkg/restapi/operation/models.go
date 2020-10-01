/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
)

type healthCheckResp struct {
	Status      string    `json:"status"`
	CurrentTime time.Time `json:"currentTime"`
}

// DIDCommInvitationResp model.
type DIDCommInvitationResp struct {
	Invitation *outofband.Invitation `json:"invitation"`
}

// EstablishConn model.
type EstablishConn struct {
	ID      string             `json:"@id"`
	Type    string             `json:"@type"`
	Purpose []string           `json:"~purpose"`
	Data    *EstablishConnData `json:"data"`
}

// EstablishConnData model for data in EstablishConn.
type EstablishConnData struct {
	DIDDoc *did.Doc `json:"didDoc"`
}
