/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package aries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
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

		_, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		expectErr := fmt.Errorf("expected error")

		ctx := ariesMockProvider()
		ctx.StorageProviderValue = &mockstore.MockStoreProvider{ErrOpenStoreHandle: expectErr}

		_, err := newPublicDIDGetter(ctx, nil)
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

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		_, err = pdg.Initialize("", "", nil, "")
		require.NoError(t, err)
	})

	t.Run("fail: create auth verification", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = "oopsie-woopsie-not-a-key-type"
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		_, err = pdg.Initialize("", "", nil, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating did doc Authentication")
	})

	t.Run("fail: create KeyAgreement verification", func(t *testing.T) {
		ctx := ariesMockProvider()
		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = "oopsie-woopsie-not-a-key-type"

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		_, err = pdg.Initialize("", "", nil, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating did doc KeyAgreement")
	})

	t.Run("fail: creating orb DID", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		_, err = pdg.Initialize("", "", nil, "")
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

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			b, e := json.Marshal(did.DocResolution{
				DIDDocument: &did.Doc{ID: "did:orb:test", Context: []string{did.ContextV1}},
			})
			require.NoError(t, e)

			_, e = w.Write(b)
			require.NoError(t, e)
		}))

		defer srv.Close()

		_, err = pdg.Initialize("", "", []string{srv.URL}, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "error saving public DID")
		require.ErrorIs(t, err, expectErr)
	})

	t.Run("success: create and save orb DID", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx = addRealKMS(t, ctx)

		ctx.KeyTypeValue = kms.ECDSAP256IEEEP1363
		ctx.KeyAgreementTypeValue = kms.NISTP256ECDHKWType

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		testDID := "did:orb:test"

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			b, e := json.Marshal(did.DocResolution{
				DIDDocument: &did.Doc{ID: testDID, Context: []string{did.ContextV1}},
			})
			require.NoError(t, e)

			_, e = w.Write(b)
			require.NoError(t, e)
		}))

		defer srv.Close()

		res, err := pdg.Initialize("", "", []string{srv.URL}, "")
		require.NoError(t, err)
		require.Equal(t, testDID, res)
	})
}

func TestPublicDIDGetter_createVerification(t *testing.T) {
	t.Run("fail: key can't be converted to jwk", func(t *testing.T) {
		ctx := ariesMockProvider()

		ctx.KMSValue = &mockkms.KeyManager{CrAndExportPubKeyValue: []byte("foo bar baz")}

		pdg, err := newPublicDIDGetter(ctx, nil)
		require.NoError(t, err)

		_, err = pdg.createVerification("foo", "foo", 0)
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
