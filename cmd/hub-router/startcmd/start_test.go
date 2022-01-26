/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type mockServer struct{}

func (m *mockServer) ListenAndServe(host string, router http.Handler) error {
	return nil
}

func (m *mockServer) ListenAndServeTLS(host, certFile, keyFile string, router http.Handler) error {
	return nil
}

func TestHTTPServer_ListenAndServeTLS(t *testing.T) {
	var w HTTPServer
	err := w.ListenAndServeTLS("wronghost", "", "", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address wronghost: missing port in address")
}

type closeFunc func()

func dummySidetree(t *testing.T) (string, closeFunc) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, err := json.Marshal(did.DocResolution{
			DIDDocument: &did.Doc{ID: "did:orb:test", Context: []string{did.ContextV1}},
		})
		require.NoError(t, err)

		_, err = w.Write(b)
		require.NoError(t, err)
	}))

	return srv.URL, func() {
		srv.Close()
	}
}

func TestGetStartCmd(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		orbDomain, closeOrb := dummySidetree(t)
		defer closeOrb()

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "mem://tests",
			"--" + datasourceTransientFlagName, "mem://tests",
			"--" + didcommV1FlagName, "true",
			"--" + orbDomainsFlagName, orbDomain,
			"--" + agentHTTPResolverFlagName, "orb@" + orbDomain,
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.NoError(t, err)
	})

	t.Run("contents", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		require.Equal(t, "start", startCmd.Use)
		require.Equal(t, "Start hub-router", startCmd.Short)
		require.Equal(t, "Start hub-router", startCmd.Long)

		checkFlagPropertiesCorrect(t, startCmd, hostURLFlagName, hostURLFlagShorthand, hostURLFlagUsage)
	})

	t.Run("test blank host url arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{"--" + hostURLFlagName, ""}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Equal(t, "host-url value is empty", err.Error())
	})

	t.Run("test missing host url arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		err := startCmd.Execute()

		require.Error(t, err)
		require.Equal(t,
			"Neither host-url (command line flag) nor HUB_ROUTER_HOST_URL (environment variable) have been set.",
			err.Error())
	})

	t.Run("test blank host env var", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		err := os.Setenv(hostURLEnvKey, "")
		require.NoError(t, err)

		err = startCmd.Execute()
		require.Error(t, err)
		require.Equal(t, "HUB_ROUTER_HOST_URL value is empty", err.Error())
	})

	t.Run("missing persistent dsn arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + datasourceTransientFlagName, "mem://tests",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Equal(t,
			"Neither dsn-p (command line flag) nor HUB_ROUTER_DSN_PERSISTENT (environment variable) have been set.", err.Error())
	})

	t.Run("missing transient dsn arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "mem://tests",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Equal(t,
			"Neither dsn-t (command line flag) nor HUB_ROUTER_DSN_TRANSIENT (environment variable) have been set.", err.Error())
	})

	t.Run("unsupported datasource driver", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "UNSUPPORTED://",
			"--" + datasourceTransientFlagName, "mem://tests",
			"--" + orbDomainsFlagName, "testnet.orb.trustbloc.local",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported storage driver")
	})

	t.Run("malformed dsn", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "malformed",
			"--" + datasourceTransientFlagName, "mem://tests",
			"--" + orbDomainsFlagName, "testnet.orb.trustbloc.local",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid dbURL malformed")
	})

	t.Run("invalid system cert flag", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "malformed",
			"--" + tlsSystemCertPoolFlagName, "mem://tests",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("invalid dsn timeout", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "mem://tests",
			"--" + datasourceTransientFlagName, "mem://tests",
			"--" + datasourceTimeoutFlagName, "invalid",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse dsn timeout")
	})

	t.Run("missing didcomm inbound host", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + datasourcePersistentFlagName, "mem://tests",
			"--" + datasourceTransientFlagName, "mem://tests",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Equal(t,
			"Neither didcomm-http-host (command line flag) nor HUB_ROUTER_DIDCOMM_HTTP_HOST "+
				"(environment variable) have been set.",
			err.Error())
	})

	t.Run("test adapter mode - store errors", func(t *testing.T) {
		parameters := &hubRouterParameters{
			datasourceParams: &datasourceParams{},
		}

		err := addHandlers(parameters, nil, nil, nil, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "init persistent storage: invalid dbURL")

		_, err = initStore("invaldidb://test", "", 10)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported storage driver: invaldidb")
	})

	t.Run("invalid use-didcomm-v1 bool", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "mem://tests",
			"--" + datasourceTransientFlagName, "mem://tests",
			"--" + didcommV1FlagName, "AAAAH-BAD-DATA",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "parsing use-didcomm-v1 flag")
	})

	t.Run("missing orb domain", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + didCommHTTPHostFlagName, randomURL(t),
			"--" + didCommWSHostFlagName, randomURL(t),
			"--" + datasourcePersistentFlagName, "mem://tests",
			"--" + datasourceTransientFlagName, "mem://tests",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "Neither orb-domain")
	})
}

func TestStartHubRouter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		orbDomain, closeOrb := dummySidetree(t)
		defer closeOrb()

		params := &hubRouterParameters{
			hostURL:   "localhost:8080",
			tlsParams: &tlsParameters{},
			datasourceParams: &datasourceParams{
				persistentURL: "mem://tests",
				transientURL:  "mem://tests",
			},
			didCommParameters: &didCommParameters{
				httpHostInternal: randomURL(t),
				wsHostInternal:   randomURL(t),
				useDIDCommV1:     true,
				keyType:          "ed25519",
				keyAgreementType: "x25519kw",
			},
			orbClientParameters: &orbClientParameters{
				domains: []string{orbDomain},
			},
		}

		err := startHubRouter(params, &mockServer{})
		require.NoError(t, err)
	})

	t.Run("missing tls configs", func(t *testing.T) {
		tlsParams := &tlsParameters{
			serveKeyPath: "/test",
		}

		params := &hubRouterParameters{
			tlsParams: tlsParams,
		}

		err := startHubRouter(params, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cert path and key path are mandatory : missing cert path")

		tlsParams.serveKeyPath = ""
		tlsParams.serveCertPath = "/test"

		err = startHubRouter(params, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cert path and key path are mandatory : missing key path")
	})

	t.Run("serve tls", func(t *testing.T) {
		params := &hubRouterParameters{
			tlsParams: &tlsParameters{
				serveKeyPath:  "/test",
				serveCertPath: "/test",
				caCerts:       []string{"test"},
			},
		}

		err := startHubRouter(params, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get root CAs")
	})

	t.Run("serve tls", func(t *testing.T) {
		params := &hubRouterParameters{
			tlsParams: &tlsParameters{
				serveKeyPath: "/test",
			},
		}

		err := serveHubRouter(params, &mockServer{}, nil)
		require.NoError(t, err)
	})

	t.Run("fail: initializing public DID", func(t *testing.T) {
		params := &hubRouterParameters{
			hostURL:   "localhost:8080",
			tlsParams: &tlsParameters{},
			datasourceParams: &datasourceParams{
				persistentURL: "mem://tests",
				transientURL:  "mem://tests",
			},
			didCommParameters: &didCommParameters{
				httpHostInternal: randomURL(t),
				wsHostInternal:   randomURL(t),
				useDIDCommV1:     false,
			},
			orbClientParameters: &orbClientParameters{},
		}

		err := startHubRouter(params, &mockServer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating public DID")
	})

	t.Run("fail: parsing resolver opts", func(t *testing.T) {
		params := &hubRouterParameters{
			hostURL:   "localhost:8080",
			tlsParams: &tlsParameters{},
			datasourceParams: &datasourceParams{
				persistentURL: "mem://tests",
				transientURL:  "mem://tests",
			},
			didCommParameters: &didCommParameters{
				httpHostInternal: randomURL(t),
				wsHostInternal:   randomURL(t),
				useDIDCommV1:     true,
				didResolvers:     []string{"error oops bad"},
			},
			orbClientParameters: &orbClientParameters{
				domains: []string{"foo"},
			},
		}

		err := startHubRouter(params, &mockServer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid http resolver options found")
	})

	t.Run("fail: creating http resolver", func(t *testing.T) {
		params := &hubRouterParameters{
			hostURL:   "localhost:8080",
			tlsParams: &tlsParameters{},
			datasourceParams: &datasourceParams{
				persistentURL: "mem://tests",
				transientURL:  "mem://tests",
			},
			didCommParameters: &didCommParameters{
				httpHostInternal: randomURL(t),
				wsHostInternal:   randomURL(t),
				useDIDCommV1:     true,
				didResolvers:     []string{"badResolver@not-a-url\x01^^"},
			},
			orbClientParameters: &orbClientParameters{
				domains: []string{"foo"},
			},
		}

		err := startHubRouter(params, &mockServer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to setup http resolver")
	})
}

func TestSupportedDatabases(t *testing.T) {
	tests := []struct {
		dbURL          string
		isErr          bool
		expectedErrMsg string
	}{
		{dbURL: "mem://test", isErr: false},
		{dbURL: "mongodb://", isErr: true, expectedErrMsg: "store init - connect to storage at mongodb://"},
		{dbURL: "random", isErr: true, expectedErrMsg: "invalid dbURL random"},
	}

	for _, test := range tests {
		_, err := initStore(test.dbURL, "hr-store", 1)

		if !test.isErr {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), test.expectedErrMsg)
		}
	}
}

func checkFlagPropertiesCorrect(t *testing.T, cmd *cobra.Command, flagName, flagShorthand, flagUsage string) {
	t.Helper()

	flag := cmd.Flag(flagName)

	require.NotNil(t, flag)
	require.Equal(t, flagName, flag.Name)
	require.Equal(t, flagShorthand, flag.Shorthand)
	require.Equal(t, flagUsage, flag.Usage)
	require.Equal(t, "", flag.Value.String())

	flagAnnotations := flag.Annotations
	require.Nil(t, flagAnnotations)
}

func randomURL(t *testing.T) string {
	t.Helper()

	p, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", p)
}
