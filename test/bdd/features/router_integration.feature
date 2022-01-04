#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@hub_router
Feature: Hub Router Integration

  Scenario: Establish Connection between Adapter and Router through Wallet
    When Wallet gets DIDComm invitation from hub-router
    Then Wallet connects with Router
    And  Wallet registers with the Router for mediation
    Then Wallet gets invitation from Adapter
    And  Wallet connects with Adapter
    Then Wallet sends establish connection request for adapter
    And  Wallet passes the details of router to adapter
    And  Adapter registers with the Router for mediation

  Scenario: Establish DIDComm V2 Connections from Router to Adapter and Wallet
    When Wallet gets DIDComm V2 invitation from hub-router
    Then Wallet connects with Router using DIDComm V2
    Then Wallet gets DIDComm V2 invitation from Adapter
    And  Wallet connects with Adapter using DIDComm V2
