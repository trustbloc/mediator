#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@mediator
Feature: Hub Router Integration

  Scenario: Establish Connection between Adapter and Router through Wallet
    When Wallet gets DIDComm invitation from mediator
    Then Wallet connects with Router
    And  Wallet registers with the Router for mediation
    Then Wallet gets invitation from Adapter
    And  Wallet connects with Adapter
    Then Wallet sends establish connection request for adapter
    And  Wallet passes the details of router to adapter
    And  Adapter registers with the Router for mediation
