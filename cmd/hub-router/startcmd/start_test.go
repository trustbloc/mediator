/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"fmt"
	"net/http"
	"os"
	"testing"

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

func TestGetStartCmd(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
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

		err := addHandlers(parameters, nil, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "init persistent storage provider with url : invalid dbURL")

		_, err = initEdgeStore("invaldidb://test", "", 10)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported storage driver: invaldidb")

		_, err = initAriesStore("invaldidb://test", "", 10)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported storage driver: invaldidb")
	})
}

func TestStartHubRouter(t *testing.T) {
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
}

func TestSupportedDatabases(t *testing.T) {
	t.Run("edge store", func(t *testing.T) {
		tests := []struct {
			dbURL          string
			isErr          bool
			expectedErrMsg string
		}{
			{dbURL: "mem://test", isErr: false},
			{dbURL: "mysql://test", isErr: true, expectedErrMsg: "edgestore init - connect to storage at test"},
			{dbURL: "random", isErr: true, expectedErrMsg: "invalid dbURL random"},
		}

		for _, test := range tests {
			_, err := initEdgeStore(test.dbURL, "hr-store", 1)

			if !test.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			}
		}
	})

	t.Run("aries store", func(t *testing.T) {
		tests := []struct {
			dbURL          string
			isErr          bool
			expectedErrMsg string
		}{
			{dbURL: "mem://test", isErr: false},
			{dbURL: "mysql://test", isErr: true, expectedErrMsg: "ariesstore init - connect to storage at test"},
			{dbURL: "random", isErr: true, expectedErrMsg: "invalid dbURL random"},
		}

		for _, test := range tests {
			_, err := initAriesStore(test.dbURL, "hr-store", 1)

			if !test.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			}
		}
	})
}

func checkFlagPropertiesCorrect(t *testing.T, cmd *cobra.Command, flagName, flagShorthand, flagUsage string) {
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
	p, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", p)
}
