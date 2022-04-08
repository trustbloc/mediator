#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@didcomm_v2
Feature: Hub Router Integration - DIDComm v2

  Scenario: Establish DIDComm V2 Connections from Router to Adapter and Wallet
    When Wallet gets DIDComm V2 invitation from mediator
    Then Wallet connects with Router using DIDComm V2
    Then Wallet gets DIDComm V2 invitation from Adapter
    And  Wallet connects with Adapter using DIDComm V2
