#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@healthcheck
Feature: Health Check

  Scenario: Status OK
    When an HTTP GET is sent to the healthcheck endpoint
    Then hub-router responds with status OK
