#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@hub_router
Feature: DID Exchange through Router

  Scenario: DID Exchange with Router
    When Wallet gets DIDComm invitation from hub-router
    Then Wallet connects with Router
