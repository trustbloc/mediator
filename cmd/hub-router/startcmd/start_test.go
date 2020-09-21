/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"fmt"
	"io/ioutil"
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
			"--" + didCommDBPathFlagName, tmpDir(t),
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
			"--" + didCommDBPathFlagName, tmpDir(t),
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid dbURL malformed")
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

func tmpDir(t *testing.T) string {
	path, err := ioutil.TempDir("", "db")
	require.NoError(t, err)

	t.Cleanup(func() {
		delErr := os.RemoveAll(path)
		require.NoError(t, delErr)
	})

	return path
}
