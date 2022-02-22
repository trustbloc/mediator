/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/noop"
	"github.com/stretchr/testify/require"
)

func TestGetPublicDID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		didValue := "did:foo:bar"

		store := mockstore.MockStore{Store: map[string]mockstore.DBEntry{
			storeDIDKey: {Value: []byte(didValue)},
		}}

		ctx := ariesMockProvider()
		ctx.StorageProviderValue = mockstore.NewCustomMockStoreProvider(&store)

		_, err := GetPublicDID(ctx, &PublicDIDConfig{})
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		expectErr := fmt.Errorf("expected error")

		ctx := ariesMockProvider()
		ctx.StorageProviderValue = &mockstore.MockStoreProvider{ErrOpenStoreHandle: expectErr}

		_, err := GetPublicDID(ctx, &PublicDIDConfig{})
		require.Error(t, err)
		require.ErrorIs(t, err, expectErr)
	})
}

func TestNewPublicDIDGetter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := getAriesCtx()

		_, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		expectErr := fmt.Errorf("expected error")

		ctx := ariesMockProvider()
		ctx.StorageProviderValue = &mockstore.MockStoreProvider{ErrOpenStoreHandle: expectErr}

		_, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.Error(t, err)
		require.ErrorIs(t, err, expectErr)
	})
}

func TestPublicDIDGetter_Initialize(t *testing.T) {
	t.Run("success - DID already created", func(t *testing.T) {
		didValue := "did:foo:bar"

		store := mockstore.MockStore{Store: map[string]mockstore.DBEntry{
			storeDIDKey: {Value: []byte(didValue)},
		}}

		ctx := ariesMockProvider()
		ctx.StorageProviderValue = mockstore.NewCustomMockStoreProvider(&store)

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.Initialize("")
		require.NoError(t, err)
	})

	t.Run("fail: create auth verification", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = "oopsie-woopsie-not-a-key-type"
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.Initialize("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating did doc Authentication")
	})

	t.Run("fail: create KeyAgreement verification", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = "oopsie-woopsie-not-a-key-type"

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.Initialize("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating did doc KeyAgreement")
	})

	t.Run("fail: creating orb DID", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.Initialize("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating public orb DID")
	})

	t.Run("fail: saving orb DID", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx = addRealKMS(t, ctx)

		expectErr := fmt.Errorf("expected error")
		ctx.StorageProviderValue = mockstore.NewCustomMockStoreProvider(&mockstore.MockStore{ErrPut: expectErr})

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		pdg.orbVDR = &mockVDR{createFunc: func(didDoc *did.Doc, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
			return &did.DocResolution{
				DIDDocument: &did.Doc{ID: "did:orb:test", Context: []string{did.ContextV1}},
			}, nil
		}}

		_, err = pdg.Initialize("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "error saving public DID")
		require.ErrorIs(t, err, expectErr)
	})

	t.Run("success: create and save orb DID", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		testDID := "did:orb:test"

		pdg.orbVDR = &mockVDR{createFunc: func(didDoc *did.Doc, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
			return &did.DocResolution{
				DIDDocument: &did.Doc{ID: testDID, Context: []string{did.ContextV1}},
			}, nil
		}}

		res, err := pdg.Initialize("")
		require.NoError(t, err)
		require.Equal(t, testDID, res)
	})
}

func TestPublicDIDGetter_createVerification(t *testing.T) {
	t.Run("fail: key can't be converted to jwk", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx.KMSValue = &mockkms.KeyManager{CrAndExportPubKeyValue: []byte("foo bar baz")}

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.createVerification("foo", "foo", 0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating jwk")
	})

	t.Run("success: dummy ed25519 key", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx.KMSValue = &mockkms.KeyManager{CrAndExportPubKeyValue: []byte("foo bar baz")}
		ctx.KeyTypeValue = kms.ED25519Type

		pdg, err := newPublicDIDGetter(ctx, nil, nil, "t1")
		require.NoError(t, err)

		_, err = pdg.createVerification("foo", kms.ED25519Type, 0)
		require.NoError(t, err)
	})
}

func TestCreateVerification(t *testing.T) {
	t.Run("fail: creating pub key", func(t *testing.T) {
		_, err := CreateVerification(&mockkms.KeyManager{CrAndExportPubKeyErr: fmt.Errorf("uh oh")},
			"#key-1", kms.ECDSAP256IEEEP1363, did.Authentication)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating public key")
	})

	t.Run("success: Ed25519 key", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		vm, err := CreateVerification(ctx.KMS(), "#key-1", kms.ED25519Type, did.Authentication)
		require.NoError(t, err)
		require.Equal(t, "#key-1", vm.VerificationMethod.ID)
		require.Equal(t, "Ed25519VerificationKey2018", vm.VerificationMethod.Type)
	})

	t.Run("success: X25519 key", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		vm, err := CreateVerification(ctx.KMS(), "#key-1", kms.X25519ECDHKWType, did.Authentication)
		require.NoError(t, err)
		require.Equal(t, "#key-1", vm.VerificationMethod.ID)
		require.Equal(t, "X25519KeyAgreementKey2019", vm.VerificationMethod.Type)
	})

	t.Run("fail: creating VM from bad data for X25519 key", func(t *testing.T) {
		_, err := CreateVerification(&mockkms.KeyManager{CrAndExportPubKeyValue: []byte("uh oh")},
			"#key-1", kms.X25519ECDHKWType, did.Authentication)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unmarshal X25519 key")
	})

	t.Run("success: ECDSA P256 key", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		vm, err := CreateVerification(ctx.KMS(), "#key-1", kms.ECDSAP256IEEEP1363, did.Authentication)
		require.NoError(t, err)
		require.Equal(t, "#key-1", vm.VerificationMethod.ID)

		jwk := vm.VerificationMethod.JSONWebKey()
		require.Equal(t, "P-256", jwk.Crv)
	})

	t.Run("fail: creating JWK from bad data for ECDSA P256 key", func(t *testing.T) {
		_, err := CreateVerification(&mockkms.KeyManager{CrAndExportPubKeyValue: []byte("uh oh")},
			"#key-1", kms.ECDSAP256IEEEP1363, did.Authentication)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating jwk")
	})
}

func addRealKMS(t *testing.T, ctx *mockprovider.Provider) *mockprovider.Provider {
	t.Helper()

	var err error

	ctx.SecretLockValue = &noop.NoLock{}

	ctx.KMSValue, err = localkms.New("prefixname://test.kms", ctx)
	require.NoError(t, err)

	return ctx
}

type mockVDR struct {
	createFunc func(didDoc *did.Doc, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error)
}

func (m *mockVDR) Create(didDoc *did.Doc, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
	if m.createFunc != nil {
		return m.createFunc(didDoc, opts...)
	}

	return nil, nil
}
