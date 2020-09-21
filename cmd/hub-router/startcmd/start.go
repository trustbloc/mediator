/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/mux"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	arieshttp "github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/http"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariesctx "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/spf13/cobra"
	"github.com/trustbloc/edge-core/pkg/log"
	"github.com/trustbloc/edge-core/pkg/storage"
	"github.com/trustbloc/edge-core/pkg/storage/memstore"
	cmdutils "github.com/trustbloc/edge-core/pkg/utils/cmd"
	tlsutils "github.com/trustbloc/edge-core/pkg/utils/tls"

	routeraries "github.com/trustbloc/hub-router/pkg/aries"
	"github.com/trustbloc/hub-router/pkg/restapi/operation"
)

// Network config.
const (
	hostURLFlagName      = "host-url"
	hostURLFlagShorthand = "u"
	hostURLFlagUsage     = "URL to run the hub-router instance on. Format: HostName:Port." +
		" Alternatively, this can be set with the following environment variable: " + hostURLEnvKey
	hostURLEnvKey = "HUB_ROUTER_HOST_URL"

	didCommInboundHostFlagName  = "didcomm-inbound-host"
	didCommInboundHostEnvKey    = "HUB_ROUTER_DIDCOMM_INBOUND_HOST"
	didCommInboundHostFlagUsage = "Inbound Host Name:Port. This is used internally to start the didcomm server." +
		" Alternatively, this can be set with the following environment variable: " + didCommInboundHostEnvKey

	// inbound host external url flag.
	didCommInboundHostExternalFlagName  = "didcomm-inbound-host-external"
	didCommInboundHostExternalEnvKey    = "HUB_ROUTER_DIDCOMM_INBOUND_HOST_EXTERNAL"
	didCommInboundHostExternalFlagUsage = "Inbound Host External Name:Port." +
		" This is the URL for the inbound server as seen externally." +
		" If not provided, then the internal inbound host will be used here." +
		" Alternatively, this can be set with the following environment variable: " + didCommInboundHostExternalEnvKey

	tlsSystemCertPoolFlagName  = "tls-systemcertpool"
	tlsSystemCertPoolFlagUsage = "Use system certificate pool." +
		" Possible values [true] [false]. Defaults to false if not set." +
		" Alternatively, this can be set with the following environment variable: " + tlsSystemCertPoolEnvKey
	tlsSystemCertPoolEnvKey = "HUB_ROUTER_TLS_SYSTEMCERTPOOL"

	tlsCACertsFlagName  = "tls-cacerts"
	tlsCACertsFlagUsage = "Comma-Separated list of ca certs path." +
		" Alternatively, this can be set with the following environment variable: " + tlsCACertsEnvKey
	tlsCACertsEnvKey = "HUB_ROUTER_TLS_CACERTS"

	tlsServeCertPathFlagName  = "tls-serve-cert"
	tlsServeCertPathFlagUsage = "Path to the server certificate to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeCertPathEnvKey
	tlsServeCertPathEnvKey = "HUB_ROUTER_TLS_SERVE_CERT"

	tlsServeKeyPathFlagName  = "tls-serve-key"
	tlsServeKeyPathFlagUsage = "Path to the private key to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeKeyPathFlagEnvKey
	tlsServeKeyPathFlagEnvKey = "HUB_ROUTER_TLS_SERVE_KEY"
)

// Storage config.
const (
	storagePrefix = "hubrouter"

	datasourcePersistentFlagName  = "dsn-p"
	datasourcePersistentFlagUsage = "Persistent datasource Name with credentials if required." +
		" Format must be <driver>:[//]<driver-specific-dsn>." +
		" Examples: 'mysql://root:secret@tcp(localhost:3306)/hubrouter', 'mem://test'." +
		" Supported drivers are [mem, mysql]." +
		" Alternatively, this can be set with the following environment variable: " + datasourcePersistentEnvKey
	datasourcePersistentEnvKey = "HUB_ROUTER_DSN_PERSISTENT"

	datasourceTransientFlagName  = "dsn-t"
	datasourceTransientFlagUsage = "Datasource Name with credentials if required." +
		" Format must be <driver>:[//]<driver-specific-dsn>." +
		" Examples: 'mysql://root:secret@tcp(localhost:3306)/hubrouter', 'mem://test'." +
		" Supported drivers are [mem, mysql]." +
		" Alternatively, this can be set with the following environment variable: " + datasourceTransientEnvKey
	datasourceTransientEnvKey = "HUB_ROUTER_DSN_TRANSIENT"

	// db path.
	didCommDBPathFlagName  = "didcomm-db-path"
	didCommDBPathEnvKey    = "HUB_ROUTER_DIDCOMM_DB_PATH"
	didCommDBPathFlagUsage = "Path to database." +
		" Alternatively, this can be set with the following environment variable: " + didCommDBPathEnvKey
)

// "Other" bucket.
const (
	logLevelFlagName  = "log-level"
	logLevelFlagUsage = "Sets the logging level." +
		" Possible values are [DEBUG, INFO, WARNING, ERROR, CRITICAL] (default is INFO)." +
		" Alternatively, this can be set with the following environment variable: " + logLevelEnvKey
	logLevelEnvKey = "HUB_ROUTER_LOGLEVEL"
)

var logger = log.New("hub-router")

// nolint:gochecknoglobals // we map the <driver> portion of datasource URLs to this map's keys
var supportedEdgeStorageProviders = map[string]func(string, string) (storage.Provider, error){
	"mem": func(_, _ string) (storage.Provider, error) { // nolint:unparam // memstorage provider never returns error
		return memstore.NewProvider(), nil
	},
}

type tlsParameters struct {
	systemCertPool bool
	caCerts        []string
	serveCertPath  string
	serveKeyPath   string
}

type didCommParameters struct {
	inboundHostInternal string
	inboundHostExternal string
}

type datasourceParams struct {
	persistentURL string
	transientURL  string
	didcommDBPath string
}

type hubRouterParameters struct {
	hostURL           string
	tlsParams         *tlsParameters
	datasourceParams  *datasourceParams
	didCommParameters *didCommParameters
}

type server interface {
	ListenAndServeTLS(host, certFile, keyFile string, router http.Handler) error
}

// HTTPServer represents an actual HTTP server implementation.
type HTTPServer struct{}

// ListenAndServeTLS starts the server using the standard Go HTTPS implementation.
func (s *HTTPServer) ListenAndServeTLS(host, certFile, keyFile string, router http.Handler) error {
	return http.ListenAndServeTLS(host, certFile, keyFile, router)
}

// GetStartCmd returns the Cobra start command.
func GetStartCmd(srv server) *cobra.Command {
	startCmd := createStartCmd(srv)

	createFlags(startCmd)

	return startCmd
}

func createStartCmd(srv server) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start hub-router",
		Long:  "Start hub-router",
		RunE: func(cmd *cobra.Command, args []string) error {
			parameters, err := getHubRouterParameters(cmd)
			if err != nil {
				return err
			}

			return startHubRouter(parameters, srv)
		},
	}
}

func createFlags(startCmd *cobra.Command) {
	startCmd.Flags().StringP(hostURLFlagName, hostURLFlagShorthand, "", hostURLFlagUsage)
	startCmd.Flags().StringP(tlsSystemCertPoolFlagName, "", "", tlsSystemCertPoolFlagUsage)
	startCmd.Flags().StringArrayP(tlsCACertsFlagName, "", []string{}, tlsCACertsFlagUsage)
	startCmd.Flags().StringP(tlsServeCertPathFlagName, "", "", tlsServeCertPathFlagUsage)
	startCmd.Flags().StringP(tlsServeKeyPathFlagName, "", "", tlsServeKeyPathFlagUsage)
	startCmd.Flags().StringP(datasourcePersistentFlagName, "", "", datasourcePersistentFlagUsage)
	startCmd.Flags().StringP(datasourceTransientFlagName, "", "", datasourceTransientFlagUsage)
	startCmd.Flags().StringP(didCommDBPathFlagName, "", "", didCommDBPathFlagUsage)

	// didcomm
	startCmd.Flags().StringP(didCommInboundHostFlagName, "", "", didCommInboundHostFlagUsage)
	startCmd.Flags().StringP(didCommInboundHostExternalFlagName, "", "", didCommInboundHostExternalFlagUsage)

	startCmd.Flags().StringP(logLevelFlagName, "", "INFO", logLevelFlagUsage)
}

func getHubRouterParameters(cmd *cobra.Command) (*hubRouterParameters, error) {
	hostURL, err := cmdutils.GetUserSetVarFromString(cmd, hostURLFlagName, hostURLEnvKey, false)
	if err != nil {
		return nil, err
	}

	tlsParams, err := getTLS(cmd)
	if err != nil {
		return nil, err
	}

	dsParams, err := getDatasourceParams(cmd)
	if err != nil {
		return nil, err
	}

	// didcomm
	didCommParameters, err := getDIDCommParams(cmd)
	if err != nil {
		return nil, err
	}

	logLevel, err := cmdutils.GetUserSetVarFromString(cmd, logLevelFlagName, logLevelEnvKey, true)
	if err != nil {
		return nil, err
	}

	if logLevel == "" {
		logLevel = "INFO"
	}

	err = setLogLevel(logLevel)
	if err != nil {
		return nil, err
	}

	logger.Infof("logger level set to %s", logLevel)

	return &hubRouterParameters{
		hostURL:           hostURL,
		tlsParams:         tlsParams,
		datasourceParams:  dsParams,
		didCommParameters: didCommParameters,
	}, nil
}

func getTLS(cmd *cobra.Command) (*tlsParameters, error) {
	tlsSystemCertPoolString, err := cmdutils.GetUserSetVarFromString(cmd, tlsSystemCertPoolFlagName,
		tlsSystemCertPoolEnvKey, true)
	if err != nil {
		return nil, err
	}

	tlsSystemCertPool := false
	if tlsSystemCertPoolString != "" {
		tlsSystemCertPool, err = strconv.ParseBool(tlsSystemCertPoolString)
		if err != nil {
			return nil, err
		}
	}

	tlsCACerts, err := cmdutils.GetUserSetVarFromArrayString(cmd, tlsCACertsFlagName, tlsCACertsEnvKey, true)
	if err != nil {
		return nil, err
	}

	tlsServeCertPath, err := cmdutils.GetUserSetVarFromString(cmd, tlsServeCertPathFlagName, tlsServeCertPathEnvKey, true)
	if err != nil {
		return nil, err
	}

	tlsServeKeyPath, err := cmdutils.GetUserSetVarFromString(cmd, tlsServeKeyPathFlagName, tlsServeKeyPathFlagEnvKey, true)
	if err != nil {
		return nil, err
	}

	return &tlsParameters{
		systemCertPool: tlsSystemCertPool,
		caCerts:        tlsCACerts,
		serveCertPath:  tlsServeCertPath,
		serveKeyPath:   tlsServeKeyPath,
	}, nil
}

func getDatasourceParams(cmd *cobra.Command) (*datasourceParams, error) {
	params := &datasourceParams{}

	var err error

	params.persistentURL, err = cmdutils.GetUserSetVarFromString(cmd,
		datasourcePersistentFlagName, datasourcePersistentEnvKey, false)
	if err != nil {
		return nil, err
	}

	params.transientURL, err = cmdutils.GetUserSetVarFromString(cmd,
		datasourceTransientFlagName, datasourceTransientEnvKey, false)
	if err != nil {
		return nil, err
	}

	params.didcommDBPath, err = cmdutils.GetUserSetVarFromString(cmd,
		didCommDBPathFlagName, didCommDBPathEnvKey, true)

	return params, err
}

func getDIDCommParams(cmd *cobra.Command) (*didCommParameters, error) {
	inboundHostInternal, err := cmdutils.GetUserSetVarFromString(cmd, didCommInboundHostFlagName,
		didCommInboundHostEnvKey, false)
	if err != nil {
		return nil, err
	}

	inboundHostExternal, err := cmdutils.GetUserSetVarFromString(cmd, didCommInboundHostExternalFlagName,
		didCommInboundHostExternalEnvKey, true)
	if err != nil {
		return nil, err
	}

	return &didCommParameters{
		inboundHostInternal: inboundHostInternal,
		inboundHostExternal: inboundHostExternal,
	}, nil
}

func setLogLevel(logLevel string) error {
	err := setEdgeCoreLogLevel(logLevel)
	if err != nil {
		return err
	}

	return setAriesFrameworkLogLevel(logLevel)
}

func setEdgeCoreLogLevel(logLevel string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level '%s' : %w", logLevel, err)
	}

	log.SetLevel("", level)

	return nil
}

func setAriesFrameworkLogLevel(logLevel string) error {
	level, err := arieslog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level '%s' : %w", logLevel, err)
	}

	arieslog.SetLevel("", level)

	return nil
}

func startHubRouter(params *hubRouterParameters, srv server) error {
	rootCAs, err := tlsutils.GetCertPool(params.tlsParams.systemCertPool, params.tlsParams.caCerts)
	if err != nil {
		return err
	}

	logger.Debugf("root ca's %v", rootCAs)

	router := mux.NewRouter()

	ariesCtx, err := createAriesAgent(params, &tls.Config{RootCAs: rootCAs, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}

	err = addHandlers(params, ariesCtx, router)
	if err != nil {
		return fmt.Errorf("failed to add handlers: %w", err)
	}

	logger.Infof("starting hub-router server on host %s", params.hostURL)

	return srv.ListenAndServeTLS(
		params.hostURL,
		params.tlsParams.serveCertPath,
		params.tlsParams.serveKeyPath,
		router,
	)
}

func addHandlers(params *hubRouterParameters, ariesCtx routeraries.Ctx, router *mux.Router) error {
	store, tStore, err := initAllEdgeStores(params.datasourceParams)
	if err != nil {
		return err
	}

	o, err := operation.New(&operation.Config{
		Aries: ariesCtx,
		Storage: &operation.Storage{
			Persistent: store,
			Transient:  tStore,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add operation handlers: %w", err)
	}

	handlers := o.GetRESTHandlers()

	for _, h := range handlers {
		router.HandleFunc(h.Path(), h.Handle()).Methods(h.Method())
	}

	return nil
}

func createAriesAgent(parameters *hubRouterParameters, tlsConfig *tls.Config) (*ariesctx.Provider, error) {
	var opts []aries.Option

	if parameters.datasourceParams.didcommDBPath != "" {
		opts = append(opts, defaults.WithStorePath(parameters.datasourceParams.didcommDBPath))
	}

	inboundTransportOpt := defaults.WithInboundHTTPAddr(
		parameters.didCommParameters.inboundHostInternal,
		parameters.didCommParameters.inboundHostExternal,
		"",
		"",
	)

	opts = append(opts, inboundTransportOpt)

	outbound, err := arieshttp.NewOutbound(arieshttp.WithOutboundTLSConfig(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("aries-framework - failed to create outbound tranpsort opts : %w", err)
	}

	opts = append(opts, aries.WithOutboundTransports(outbound))

	framework, err := aries.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("aries-framework - failed to initialize framework : %w", err)
	}

	ctx, err := framework.Context()
	if err != nil {
		return nil, fmt.Errorf("aries-framework - failed to get aries context : %w", err)
	}

	return ctx, nil
}

func initAllEdgeStores(params *datasourceParams) (persistent, transient storage.Provider, err error) {
	persistent, err = initEdgeStore(params.persistentURL, storagePrefix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init persistent storage provider with url %s: %w",
			params.persistentURL, err)
	}

	transient, err = initEdgeStore(params.transientURL, storagePrefix+"_txn")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init transient storage provider with url %s: %w",
			params.transientURL, err)
	}

	return persistent, transient, nil
}

func initEdgeStore(dbURL, prefix string) (storage.Provider, error) {
	const (
		sleep      = 1 * time.Second
		numRetries = 30
		urlParts   = 2
	)

	parsed := strings.SplitN(dbURL, ":", urlParts)

	if len(parsed) != urlParts {
		return nil, fmt.Errorf("invalid dbURL %s", dbURL)
	}

	driver := parsed[0]
	dsn := strings.TrimPrefix(parsed[1], "//")

	providerFunc, supported := supportedEdgeStorageProviders[driver]
	if !supported {
		return nil, fmt.Errorf("unsupported storage driver: %s", driver)
	}

	var store storage.Provider

	err := backoff.RetryNotify(
		func() error {
			var openErr error
			store, openErr = providerFunc(dsn, prefix)

			return openErr
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(sleep), numRetries),
		func(retryErr error, t time.Duration) {
			logger.Warnf(
				"failed to connect to storage, will sleep for %s before trying again : %s\n",
				t, retryErr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage at %s : %w", dsn, err)
	}

	return store, nil
}
