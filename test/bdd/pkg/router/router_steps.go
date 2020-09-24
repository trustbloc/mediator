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
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	didexcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/didexchange"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/hub-router/test/bdd/pkg/bddutil"
	"github.com/trustbloc/hub-router/test/bdd/pkg/context"
)

var logger = log.New("hub-router/routersteps")

const (
	hubRouterURL = "https://localhost:10200"
	walletAPIURL = "https://localhost:10210"

	receiveInvtiationPath  = "/connections/receive-invitation"
	connectionsByIDPathFmt = "/connections/%s"
)

// Steps is steps for VC BDD tests.
type Steps struct {
	bddContext    *context.BDDContext
	invitationStr string
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
	s.Step(`^Wallet connects with Router$`, e.connect)
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

	var invitation didexchange.Invitation

	err = json.Unmarshal(respBytes, &invitation)
	if err != nil {
		return fmt.Errorf("get invitation - marshal response :%w", err)
	}

	if invitation.Type != "https://didcomm.org/didexchange/1.0/invitation" {
		return fmt.Errorf("invalid invitation type : expected=%s actual=%s",
			"https://didcomm.org/didexchange/1.0/invitation", invitation.Type)
	}

	e.invitationStr = string(respBytes)

	return nil
}

func (e *Steps) connect() error {
	// receive invitation
	connID, err := e.receiveInvitation()
	if err != nil {
		return fmt.Errorf("receive inviation : %w", err)
	}

	// verify the connection
	err = e.validateConnection(connID, "completed")
	if err != nil {
		return fmt.Errorf("validate connection : %w", err)
	}

	return nil
}

func (e *Steps) receiveInvitation() (string, error) {
	resp, err := bddutil.HTTPDo(http.MethodPost, //nolint:bodyclose // false positive as body is closed in util function
		walletAPIURL+receiveInvtiationPath, "", "", bytes.NewBufferString(e.invitationStr), e.bddContext.TLSConfig)
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

	var connRes didexcmd.ReceiveInvitationResponse

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
