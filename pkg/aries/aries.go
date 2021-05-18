/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	ariescrypto "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/spi/storage"
)

// Ctx framework context provider.
type Ctx interface {
	Service(id string) (interface{}, error)
	ServiceEndpoint() string
	RouterEndpoint() string
	StorageProvider() storage.Provider
	ProtocolStateStorageProvider() storage.Provider
	KMS() kms.KeyManager
	VDRegistry() vdrapi.Registry
	Crypto() ariescrypto.Crypto
	KeyType() kms.KeyType
	KeyAgreementType() kms.KeyType
}

// OutOfBand client.
type OutOfBand interface {
	CreateInvitation(services []interface{}, opts ...outofband.MessageOption) (*outofband.Invitation, error)
}

// DIDExchange client.
type DIDExchange interface {
	CreateConnection(myDID string, theirDID *did.Doc, options ...didexchange.ConnectionOption) (string, error)
	RegisterActionEvent(chan<- service.DIDCommAction) error
	GetConnection(connectionID string) (*didexchange.Connection, error)
}

// Mediator client.
type Mediator interface {
	RegisterActionEvent(chan<- service.DIDCommAction) error
}

// CreateOutofbandClient util function to create oob client.
func CreateOutofbandClient(ariesCtx outofband.Provider) (*outofband.Client, error) {
	oobClient, err := outofband.New(ariesCtx)
	if err != nil {
		return nil, fmt.Errorf("create out-of-band client : %w", err)
	}

	return oobClient, err
}

// CreateDIDExchangeClient util function to create did exchange client and registers for action event.
func CreateDIDExchangeClient(ctx Ctx, actionCh chan service.DIDCommAction,
	stateMsgCh chan service.StateMsg) (DIDExchange, error) {
	didExClient, err := didexchange.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create didexchange client : %w", err)
	}

	err = didExClient.RegisterActionEvent(actionCh)
	if err != nil {
		return nil, fmt.Errorf("register didexchange action event : %w", err)
	}

	err = didExClient.RegisterMsgEvent(stateMsgCh)
	if err != nil {
		return nil, fmt.Errorf("register didexchange message event : %w", err)
	}

	return didExClient, nil
}

// CreateMediatorClient util function to create mediator client and registers for action event.
func CreateMediatorClient(ctx Ctx, actionCh chan service.DIDCommAction) (Mediator, error) {
	mediatorClient, err := mediator.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create mediator client : %w", err)
	}

	err = mediatorClient.RegisterActionEvent(actionCh)
	if err != nil {
		return nil, fmt.Errorf("register mediator action event : %w", err)
	}

	return mediatorClient, nil
}
