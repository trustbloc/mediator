/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	mockcrypto "github.com/hyperledger/aries-framework-go/pkg/mock/crypto"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	mockvdri "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"golang.org/x/crypto/curve25519"

	"github.com/trustbloc/mediator/pkg/aries"
	mockoutofband "github.com/trustbloc/mediator/pkg/internal/mock/outofband"
	mockoutofbandv2 "github.com/trustbloc/mediator/pkg/internal/mock/outofbandv2"
)

func getAriesCtx() aries.Ctx {
	return getMockProvider()
}

func getMockProvider() *mockprovider.Provider {
	x25519key, _ := curve25519.X25519( // nolint:errcheck // test helper.
		make([]byte, curve25519.ScalarSize),
		curve25519.Basepoint,
	)

	key := &crypto.PublicKey{
		X: x25519key,
	}

	keyBytes, _ := json.Marshal(key) // nolint:errcheck // test helper.

	return &mockprovider.Provider{
		ProtocolStateStorageProviderValue: mockstore.NewMockStoreProvider(),
		StorageProviderValue:              mockstore.NewMockStoreProvider(),
		ServiceMap: map[string]interface{}{
			outofbandsvc.Name:       &mockoutofband.MockService{},
			outofbandv2.Name:        &mockoutofbandv2.MockService{},
			didexchange.DIDExchange: &mocksvc.MockDIDExchangeSvc{},
			mediator.Coordination:   &mockroute.MockMediatorSvc{},
		},
		KMSValue: &mockkms.KeyManager{
			CrAndExportPubKeyValue: keyBytes,
			ImportPrivateKeyErr:    fmt.Errorf("error import priv key"),
		},
		CryptoValue:           &mockcrypto.Crypto{},
		ServiceEndpointValue:  "endpoint",
		VDRegistryValue:       &mockvdri.MockVDRegistry{},
		KeyTypeValue:          kms.ED25519Type,
		KeyAgreementTypeValue: kms.X25519ECDHKWType,
	}
}

func config(ctx ...aries.Ctx) *Config {
	var ariesCtx aries.Ctx

	if len(ctx) > 0 {
		ariesCtx = ctx[0]
	} else {
		ariesCtx = getAriesCtx()
	}

	return &Config{
		Aries:        ariesCtx,
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
