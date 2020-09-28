/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cucumber/godog"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	didexcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/didexchange"
	oobcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/outofband"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/hub-router/pkg/restapi/operation"
	"github.com/trustbloc/hub-router/test/bdd/pkg/bddutil"
	"github.com/trustbloc/hub-router/test/bdd/pkg/context"
)

var logger = log.New("hub-router/routersteps")

const (
	hubRouterURL  = "https://localhost:10200"
	walletAPIURL  = "https://localhost:10210"
	adapterAPIURL = "https://localhost:10220"

	createInvitationPath   = "/outofband/create-invitation"
	acceptInvitationPath   = "/outofband/accept-invitation"
	connectionsByIDPathFmt = "/connections/%s"
)

// Steps is steps for VC BDD tests.
type Steps struct {
	bddContext           *context.BDDContext
	routerInvitationStr  *outofband.Invitation
	routerConnID         string
	adapterInvitationStr *outofband.Invitation
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

	if result.Invitation.Type != "https://didcomm.org/oob-invitation/1.0/invitation" {
		return fmt.Errorf("invalid invitation type : expected=%s actual=%s",
			"https://didcomm.org/oob-invitation/1.0/invitation", result.Invitation.Type)
	}

	e.routerInvitationStr = result.Invitation

	return nil
}

func (e *Steps) connectWithRouter() error {
	err := e.connect(e.routerInvitationStr)
	if err != nil {
		return fmt.Errorf("connect with router : %w", err)
	}

	return nil
}

func (e *Steps) connectWithAdapter() error {
	err := e.connect(e.adapterInvitationStr)
	if err != nil {
		return fmt.Errorf("connect with adapter : %w", err)
	}

	return nil
}

func (e *Steps) connect(invitation *outofband.Invitation) error {
	// receive invitation
	connID, err := e.receiveInvitation(invitation)
	if err != nil {
		return fmt.Errorf("receive inviation : %w", err)
	}

	// verify the connection
	err = e.validateConnection(connID, "completed")
	if err != nil {
		return fmt.Errorf("validate connection : %w", err)
	}

	e.routerConnID = connID

	return nil
}

func (e *Steps) mediationRegistration() error {
	reqBytes, err := json.Marshal(struct {
		ConnectionID string `json:"connectionID"`
	}{ConnectionID: e.routerConnID})
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

	if result.Invitation.Type != "https://didcomm.org/oob-invitation/1.0/invitation" {
		return fmt.Errorf("invalid invitation type : expected=%s actual=%s",
			"https://didcomm.org/oob-invitation/1.0/invitation", result.Invitation.Type)
	}

	e.adapterInvitationStr = result.Invitation

	return nil
}

func (e *Steps) receiveInvitation(invitation *outofband.Invitation) (string, error) {
	req := oobcmd.AcceptInvitationArgs{
		Invitation: invitation,
		MyLabel:    "wallet",
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
