/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package healthcheck

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cucumber/godog"

	bddctx "github.com/trustbloc/mediator/test/bdd/pkg/context"
)

const (
	hubRouterBaseURL    = "https://localhost:10200"
	healthCheckEndpoint = hubRouterBaseURL + "/healthcheck"
)

// Steps for the BDD tests.
type Steps struct {
	healchCheckResult int
	context           *bddctx.BDDContext
}

// NewSteps returns a new Steps.
func NewSteps(ctx *bddctx.BDDContext) *Steps {
	return &Steps{context: ctx}
}

// RegisterSteps registers the BDD test steps in the bdd test suite.
func (s *Steps) RegisterSteps(g *godog.Suite) {
	g.Step(`an HTTP GET is sent to the healthcheck endpoint`, s.requestHealthCheck)
	g.Step(`mediator responds with status OK`, s.confirmHealthResult)
}

func (s *Steps) requestHealthCheck() error {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: s.context.TLSConfig}}
	defer client.CloseIdleConnections()

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, healthCheckEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to build healthcheck request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to HTTP GET endpoint: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			fmt.Printf("warning - failed to close HTTP response body: %s", err.Error())
		}
	}()

	s.healchCheckResult = resp.StatusCode

	return nil
}

func (s *Steps) confirmHealthResult() error {
	if s.healchCheckResult != http.StatusOK {
		return fmt.Errorf("expected %d but got %d", http.StatusOK, s.healchCheckResult)
	}

	return nil
}
