/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
)

type healthCheckResp struct {
	Status      string    `json:"status"`
	CurrentTime time.Time `json:"currentTime"`
}

// DIDCommInvitationResp model.
type DIDCommInvitationResp struct {
	Invitation *outofband.Invitation `json:"invitation"`
}

// CreateConnReq model.
type CreateConnReq struct {
	ID      string             `json:"@id"`
	Type    string             `json:"@type"`
	Purpose []string           `json:"~purpose"`
	Data    *CreateConnReqData `json:"data"`
}

// CreateConnReqData model for data in CreateConnReq.
type CreateConnReqData struct {
	DIDDoc []byte `json:"thirdPartyDIDDoc"`
}

// CreateConnResp model.
type CreateConnResp struct {
	ID      string              `json:"@id"`
	Type    string              `json:"@type"`
	Purpose []string            `json:"~purpose"`
	Data    *CreateConnRespData `json:"data"`
}

// CreateConnRespData model for error data in CreateConnResp.
type CreateConnRespData struct {
	ErrorMsg string `json:"errorMsg"`
	DIDDoc   []byte `json:"routerDIDDoc"`
}
