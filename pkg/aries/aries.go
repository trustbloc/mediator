/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariescrypto "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
)

// Ctx framework context provider.
type Ctx interface {
	Service(id string) (interface{}, error)
	ServiceEndpoint() string
	StorageProvider() storage.Provider
	ProtocolStateStorageProvider() storage.Provider
	KMS() kms.KeyManager
	VDRIRegistry() vdriapi.Registry
	Crypto() ariescrypto.Crypto
}

// DIDExchange client.
type DIDExchange interface {
	CreateInvitation(label string) (*didexchange.Invitation, error)
	RegisterActionEvent(chan<- service.DIDCommAction) error
}

// CreateDIDExchangeClient util function to create did exchange client and registers for action event.
func CreateDIDExchangeClient(ctx Ctx, actionCh chan service.DIDCommAction) (DIDExchange, error) {
	didExClient, err := didexchange.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create didexchange client : %w", err)
	}

	err = didExClient.RegisterActionEvent(actionCh)
	if err != nil {
		return nil, fmt.Errorf("register didexchange action event : %w", err)
	}

	return didExClient, nil
}
