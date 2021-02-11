/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package router

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	didexcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/controller/command/messaging"
	oobcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/outofband"
	vdrcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/hub-router/pkg/restapi/operation"
	"github.com/trustbloc/hub-router/test/bdd/pkg/bddutil"
	"github.com/trustbloc/hub-router/test/bdd/pkg/context"
)

var logger = log.New("hub-router/routersteps")

const (
	// base urls.
	hubRouterURL     = "https://localhost:10200"
	walletAPIURL     = "https://localhost:10210"
	adapterAPIURL    = "https://localhost:10220"
	walletWebhookURL = "http://localhost:10211"

	// connection paths.
	createInvitationPath   = "/outofband/create-invitation"
	acceptInvitationPath   = "/outofband/accept-invitation"
	connectionsByIDPathFmt = "/connections/%s"
	createConnectionPath   = "/connections/create"

	// msg service paths.
	msgServiceOperationID = "/message"
	msgServiceList        = msgServiceOperationID + "/services"
	registerMsgService    = msgServiceOperationID + "/register-service"
	unregisterMsgService  = msgServiceOperationID + "/unregister-service"
	sendNewMsg            = msgServiceOperationID + "/send"

	// vdr paths.
	vdrOperationID = "/vdr"
	vdrDIDPath     = vdrOperationID + "/did"
	resolveDIDPath = vdrDIDPath + "/resolve/%s"

	// webhook.
	checkForTopics               = "/checktopics"
	pullTopicsWaitInMilliSec     = 200
	pullTopicsAttemptsBeforeFail = 500 / pullTopicsWaitInMilliSec
)

// Steps is steps for VC BDD tests.
type Steps struct {
	bddContext           *context.BDDContext
	routerInvitationStr  *outofband.Invitation
	adapterInvitationStr *outofband.Invitation
	walletRouterConnID   string
	walletAdapterConnID  string
	adapterRouterConnID  string
	adapterDID           string
	routerDIDDoc         *did.Doc
}

// NewSteps returns new agent from client SDK.
func NewSteps(ctx *context.BDDContext) *Steps {
	return &Steps{
		bddContext: ctx,
	}
}

// RegisterSteps registers agent steps.
func (e *Steps) RegisterSteps(s *godog.Suite) {
	s.Step(`^Wallet gets DIDComm invitation from hub-router$`, e.invitation)
	s.Step(`^Wallet connects with Router$`, e.connectWithRouter)
	s.Step(`^Wallet registers with the Router for mediation$`, e.mediationRegistration)
	s.Step(`^Wallet gets invitation from Adapter$`, e.adapterInvitation)
	s.Step(`^Wallet connects with Adapter$`, e.connectWithAdapter)
	s.Step(`^Wallet sends establish connection request for adapter$`, e.establishConnReq)
	s.Step(`^Wallet passes the details of router to adapter$`, e.adapterEstablishConn)
	s.Step(`^Adapter registers with the Router for mediation$`, e.routeRegistration)
}

func (e *Steps) invitation() error {
	resp, err := bddutil.HTTPDo(http.MethodGet, //nolint:bodyclose // false positive as body is closed in util function
		hubRouterURL+"/didcomm/invitation", "", "", nil, e.bddContext.TLSConfig)
	if err != nil {
		return err
	}

	defer bddutil.CloseResponseBody(resp.Body)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("get invitation - read response : %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return bddutil.ExpectedStatusCodeError(http.StatusOK, resp.StatusCode, respBytes)
	}

	var result *operation.DIDCommInvitationResp

	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return fmt.Errorf("get invitation - marshal response :%w", err)
	}

	if result.Invitation.Type != "https://didcomm.org/out-of-band/1.0/invitation" {
		return fmt.Errorf("invalid invitation type : expected=%s actual=%s",
			"https://didcomm.org/out-of-band/1.0/invitation", result.Invitation.Type)
	}

	e.routerInvitationStr = result.Invitation

	return nil
}

func (e *Steps) connectWithRouter() error {
	msgSvcName := uuid.New().String()

	err := e.registerMsgServices(walletAPIURL, msgSvcName, "https://trustbloc.dev/didexchange/1.0/state-complete")
	if err != nil {
		return err
	}

	connID, err := e.connect(e.routerInvitationStr, "")
	if err != nil {
		return fmt.Errorf("connect with router : %w", err)
	}

	e.walletRouterConnID = connID

	err = e.getDIDExStateCompResp(walletWebhookURL, msgSvcName)
	if err != nil {
		return err
	}

	return nil
}

func (e *Steps) connectWithAdapter() error {
	connID, err := e.connect(e.adapterInvitationStr, e.walletRouterConnID)
	if err != nil {
		return fmt.Errorf("connect with adapter : %w", err)
	}

	e.walletAdapterConnID = connID

	conn, err := e.getConnection(walletAPIURL, connID)
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}

	e.adapterDID = conn.TheirDID

	return nil
}

func (e *Steps) connect(invitation *outofband.Invitation, routerConnID string) (string, error) {
	// receive invitation
	connID, err := e.receiveInvitation(invitation, routerConnID)
	if err != nil {
		return "", fmt.Errorf("receive inviation : %w", err)
	}

	// verify the connection
	err = e.validateConnection(connID, "completed")
	if err != nil {
		return "", fmt.Errorf("validate connection : %w", err)
	}

	return connID, nil
}

func (e *Steps) mediationRegistration() error {
	reqBytes, err := json.Marshal(struct {
		ConnectionID string `json:"connectionID"`
	}{ConnectionID: e.walletRouterConnID})
	if err != nil {
		return err
	}

	err = bddutil.SendHTTPReq(http.MethodPost, walletAPIURL+"/mediator/register", reqBytes, nil, e.bddContext.TLSConfig)
	if err != nil {
		return err
	}

	return nil
}

func (e *Steps) adapterInvitation() error {
	resp, err := bddutil.HTTPDo(http.MethodPost, //nolint:bodyclose // false positive as body is closed in util function
		adapterAPIURL+createInvitationPath, "", "", bytes.NewBufferString("{}"), e.bddContext.TLSConfig)
	if err != nil {
		return err
	}

	defer bddutil.CloseResponseBody(resp.Body)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("get invitation - read response : %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return bddutil.ExpectedStatusCodeError(http.StatusOK, resp.StatusCode, respBytes)
	}

	var result oobcmd.CreateInvitationResponse

	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return fmt.Errorf("get invitation - marshal response :%w", err)
	}

	if result.Invitation.Type != "https://didcomm.org/out-of-band/1.0/invitation" {
		return fmt.Errorf("invalid invitation type : expected=%s actual=%s",
			"https://didcomm.org/out-of-band/1.0/invitation", result.Invitation.Type)
	}

	e.adapterInvitationStr = result.Invitation

	return nil
}

func (e *Steps) receiveInvitation(invitation *outofband.Invitation, routerConnID string) (string, error) {
	req := oobcmd.AcceptInvitationArgs{
		Invitation:        invitation,
		MyLabel:           "wallet",
		RouterConnections: routerConnID,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := bddutil.HTTPDo(http.MethodPost, //nolint:bodyclose // false positive as body is closed in util function
		walletAPIURL+acceptInvitationPath, "", "", bytes.NewBuffer(reqBytes), e.bddContext.TLSConfig)
	if err != nil {
		return "", err
	}

	defer bddutil.CloseResponseBody(resp.Body)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", bddutil.ExpectedStatusCodeError(http.StatusOK, resp.StatusCode, respBytes)
	}

	var connRes oobcmd.AcceptInvitationResponse

	err = json.Unmarshal(respBytes, &connRes)
	if err != nil {
		return "", err
	}

	return connRes.ConnectionID, nil
}

func (e *Steps) establishConnReq() error {
	msgSvcName := uuid.New().String()

	// register for message service
	err := e.registerMsgServices(walletAPIURL, msgSvcName, "https://trustbloc.dev/blinded-routing/1.0/create-conn-resp")
	if err != nil {
		return err
	}

	// get adapters DID, that would be passed in request to router in the request
	conn, err := e.getConnection(walletAPIURL, e.walletAdapterConnID)
	if err != nil {
		return err
	}

	didDocument, err := e.resolveDID(adapterAPIURL, conn.TheirDID)
	if err != nil {
		return err
	}

	// send message
	err = e.sendCreateConnReq(walletAPIURL, didDocument)
	if err != nil {
		return fmt.Errorf("failed to send message : %w", err)
	}

	// get the response
	doc, err := e.getCreateConnResp(walletWebhookURL, msgSvcName)
	if err != nil {
		return fmt.Errorf("parse router did document: %w", err)
	}

	e.routerDIDDoc = doc

	return nil
}

func (e *Steps) adapterEstablishConn() error {
	connID, err := e.createConnection(adapterAPIURL, e.adapterDID, "my-label", e.routerDIDDoc)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}

	e.adapterRouterConnID = connID

	return nil
}

func (e *Steps) routeRegistration() error {
	reqBytes, err := json.Marshal(struct {
		ConnectionID string `json:"connectionID"`
	}{ConnectionID: e.adapterRouterConnID})
	if err != nil {
		return err
	}

	err = bddutil.SendHTTPReq(http.MethodPost, adapterAPIURL+"/mediator/register", reqBytes, nil, e.bddContext.TLSConfig)
	if err != nil {
		return err
	}

	return nil
}

func (e *Steps) validateConnection(connID, state string) error {
	const (
		sleep      = 1 * time.Second
		numRetries = 30
	)

	return backoff.RetryNotify(
		func() error {
			var openErr error

			var result didexcmd.QueryConnectionResponse
			if err := bddutil.SendHTTPReq(http.MethodGet, walletAPIURL+fmt.Sprintf(connectionsByIDPathFmt, connID),
				nil, &result, e.bddContext.TLSConfig); err != nil {
				return err
			}

			if result.Result.State != state {
				return fmt.Errorf("expected=%s actual=%s", state, result.Result.State)
			}

			return openErr
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(sleep), uint64(numRetries)),
		func(retryErr error, t time.Duration) {
			logger.Warnf(
				"validate connection : sleeping for %s before trying again : %s\n",
				t, retryErr)
		},
	)
}

func (e *Steps) sendCreateConnReq(controllerURL string, didDocument *did.Doc) error {
	didDocJSON, err := didDocument.JSONBytes()
	if err != nil {
		return err
	}

	msg := &operation.CreateConnReq{
		ID:   uuid.New().String(),
		Type: "https://trustbloc.dev/blinded-routing/1.0/create-conn-req",
		Data: &operation.CreateConnReqData{
			DIDDoc: json.RawMessage(didDocJSON),
		},
	}

	rawBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to get raw message bytes:  %w", err)
	}

	request := &messaging.SendNewMessageArgs{
		ConnectionID: e.walletRouterConnID,
		MessageBody:  rawBytes,
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// call controller to send message
	err = bddutil.SendHTTPReq(http.MethodPost, controllerURL+sendNewMsg, reqBytes, nil, e.bddContext.TLSConfig)
	if err != nil {
		return fmt.Errorf("failed to send message : %w", err)
	}

	return nil
}

func (e *Steps) getDIDExStateCompResp(controllerURL, msgSvcName string) error {
	_, err := e.pullMsgFromWebhookURL(controllerURL, msgSvcName)
	if err != nil {
		return fmt.Errorf("failed to pull incoming message from webhook : %w", err)
	}

	return nil
}

func (e *Steps) getCreateConnResp(controllerURL, msgSvcName string) (*did.Doc, error) {
	webhookMsg, err := e.pullMsgFromWebhookURL(controllerURL, msgSvcName)
	if err != nil {
		return nil, fmt.Errorf("failed to pull incoming message from webhook : %w", err)
	}

	// validate the response
	var message struct {
		Message operation.CreateConnResp `json:"message"`
	}

	err = webhookMsg.Decode(&message)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	if message.Message.Data == nil {
		return nil, errors.New("no data received from the router")
	}

	if message.Message.Data.ErrorMsg != "" {
		return nil, fmt.Errorf("error received from the route : %s", message.Message.Data.ErrorMsg)
	}

	if message.Message.Data.DIDDoc == nil {
		return nil, errors.New("no did document received from the router")
	}

	doc, err := did.ParseDocument(message.Message.Data.DIDDoc)
	if err != nil {
		return nil, fmt.Errorf("parse router did document: %w", err)
	}

	return doc, nil
}

func (e *Steps) pullMsgFromWebhookURL(webhookURL, topic string) (*service.DIDCommMsgMap, error) {
	var incoming struct {
		ID      string                `json:"id"`
		Topic   string                `json:"topic"`
		Message service.DIDCommMsgMap `json:"message"`
	}

	// try to pull recently pushed topics from webhook
	for i := 0; i < pullTopicsAttemptsBeforeFail; {
		err := bddutil.SendHTTPReq(http.MethodGet, webhookURL+checkForTopics,
			nil, &incoming, e.bddContext.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed pull topics from webhook, cause : %w", err)
		}

		if incoming.Topic != topic {
			continue
		}

		if len(incoming.Message) > 0 {
			return &incoming.Message, nil
		}

		i++

		time.Sleep(pullTopicsWaitInMilliSec * time.Millisecond)
	}

	return nil, fmt.Errorf("exhausted all [%d] attempts to pull topic from webhook", pullTopicsAttemptsBeforeFail)
}

func (e *Steps) resolveDID(controller, didID string) (*did.Doc, error) {
	destination := fmt.Sprintf(controller+resolveDIDPath, base64.StdEncoding.EncodeToString([]byte(didID)))

	var resp vdrcmd.Document

	err := bddutil.SendHTTPReq(http.MethodGet, destination, nil, &resp, e.bddContext.TLSConfig)
	if err != nil {
		return nil, fmt.Errorf("%s failed to fetch did=%s : %w", controller, didID, err)
	}

	doc, err := did.ParseDocument(resp.DID)
	if err != nil {
		return nil, fmt.Errorf("%s failed to parse did document : %w", controller, err)
	}

	return doc, nil
}

func (e *Steps) getConnection(controller, connectionID string) (*didexchange.Connection, error) {
	var response didexcmd.QueryConnectionResponse

	err := bddutil.SendHTTPReq(http.MethodGet, controller+fmt.Sprintf(connectionsByIDPathFmt, connectionID),
		nil, &response, e.bddContext.TLSConfig)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (e *Steps) createConnection(controllerURL, myDID, label string, theirDID *did.Doc) (string, error) {
	theirDIDBytes, err := theirDID.JSONBytes()
	if err != nil {
		return "", fmt.Errorf("theirDID failed to marshal to bytes : %w", err)
	}

	request, err := json.Marshal(&didexcmd.CreateConnectionRequest{
		MyDID: myDID,
		TheirDID: didexcmd.DIDDocument{
			ID:       theirDID.ID,
			Contents: theirDIDBytes,
		},
		TheirLabel: label,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal save connection request : %w", err)
	}

	var resp didexcmd.ConnectionIDArg

	err = bddutil.SendHTTPReq(http.MethodPost, controllerURL+createConnectionPath, request, &resp, e.bddContext.TLSConfig)
	if err != nil {
		return "", fmt.Errorf("%s failed to create connection : %w", controllerURL, err)
	}

	return resp.ID, nil
}

func (e *Steps) registerMsgServices(controllerURL, msgSvcName, msgType string) error {
	// unregister all the msg services (to clear older data)
	err := e.unregisterAllMsgServices(controllerURL)
	if err != nil {
		return err
	}

	// register create conn msg service
	params := messaging.RegisterMsgSvcArgs{
		Name: msgSvcName,
		Type: msgType,
	}

	reqBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	err = bddutil.SendHTTPReq(http.MethodPost, controllerURL+registerMsgService, reqBytes, nil, e.bddContext.TLSConfig)
	if err != nil {
		return err
	}

	// verify if the msg service created successfully
	result, err := e.getServicesList(controllerURL)
	if err != nil {
		return err
	}

	var found bool

	for _, svcName := range result {
		if svcName == msgSvcName {
			found = true

			break
		}
	}

	if !found {
		return fmt.Errorf("registered service not found : name=%s", msgSvcName)
	}

	return nil
}

func (e *Steps) getServicesList(controllerURL string) ([]string, error) {
	result := &messaging.RegisteredServicesResponse{}

	err := bddutil.SendHTTPReq(http.MethodGet, controllerURL+msgServiceList, nil, result, e.bddContext.TLSConfig)
	if err != nil {
		return nil, fmt.Errorf("get message service list : %w", err)
	}

	return result.Names, nil
}

func (e *Steps) unregisterAllMsgServices(controllerURL string) error {
	svcNames, err := e.getServicesList(controllerURL)
	if err != nil {
		return fmt.Errorf("unregister message services : %w", err)
	}

	for _, svcName := range svcNames {
		params := messaging.UnregisterMsgSvcArgs{
			Name: svcName,
		}

		reqBytes, err := json.Marshal(params)
		if err != nil {
			return err
		}

		err = bddutil.SendHTTPReq(http.MethodPost, controllerURL+unregisterMsgService, reqBytes, nil, e.bddContext.TLSConfig)
		if err != nil {
			return fmt.Errorf("unregister message services : %w", err)
		}
	}

	return nil
}
