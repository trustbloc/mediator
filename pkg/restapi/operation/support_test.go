/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
	mockcrypto "github.com/hyperledger/aries-framework-go/pkg/mock/crypto"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	mockvdri "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"

	"github.com/trustbloc/hub-router/pkg/aries"
	mockoutofband "github.com/trustbloc/hub-router/pkg/internal/mock/outofband"
	mockoutofbandv2 "github.com/trustbloc/hub-router/pkg/internal/mock/outofbandv2"
)

func getAriesCtx() aries.Ctx {
	return &mockprovider.Provider{
		ProtocolStateStorageProviderValue: mockstore.NewMockStoreProvider(),
		StorageProviderValue:              mockstore.NewMockStoreProvider(),
		ServiceMap: map[string]interface{}{
			outofbandsvc.Name:       &mockoutofband.MockService{},
			outofbandv2.Name:        &mockoutofbandv2.MockService{},
			didexchange.DIDExchange: &mocksvc.MockDIDExchangeSvc{},
			mediator.Coordination:   &mockroute.MockMediatorSvc{},
		},
		KMSValue:             &mockkms.KeyManager{ImportPrivateKeyErr: fmt.Errorf("error import priv key")},
		CryptoValue:          &mockcrypto.Crypto{},
		ServiceEndpointValue: "endpoint",
		VDRegistryValue:      &mockvdri.MockVDRegistry{},
	}
}

func config() *Config {
	return &Config{
		Aries:        getAriesCtx(),
		MsgRegistrar: msghandler.NewRegistrar(),
		Storage: &Storage{
			Persistent: mem.NewProvider(),
			Transient:  mem.NewProvider(),
		},
	}
}

type didexchangeEvent struct {
	connID    string
	invID     string
	invIDFunc func() string
}

func (d *didexchangeEvent) ConnectionID() string {
	return d.connID
}

func (d *didexchangeEvent) InvitationID() string {
	if d.invIDFunc != nil {
		return d.invIDFunc()
	}

	return d.invID
}

func (d *didexchangeEvent) All() map[string]interface{} {
	return make(map[string]interface{})
}
