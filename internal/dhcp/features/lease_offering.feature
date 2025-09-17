Feature: Lease Offering and Acceptance
  As a DHCP server
  I want to offer IP leases to clients and allow them to accept
  So that clients can obtain network connectivity

  Scenario: Offer and accept a lease
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.95"
    When I request the next available lease for client "test-client"
    Then I should receive a lease
    And I should receive a lease with IP "10.0.0.1"
    And the lease state should be "offered"
    And the lease client ID should be "test-client"
    When I accept the lease for "1h"
    Then the lease state should be "active"

  Scenario: Request same lease twice
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.95"
    When I request the next available lease for client "test-client"
    And I request the next available lease for client "test-client"
    Then I should receive a lease with IP "10.0.0.1"
    And the lease state should be "offered"
    When I accept the lease for "1h"
    Then the lease state should be "active"

  Scenario: Request lease after expiry for new client
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.95"
    When I request the next available lease for client "old-client"
    And the lease for client "old-client" expires in the past
    And I request the next available lease for client "new-client"
    Then I should receive a lease with IP "10.0.0.1"
    And the lease client ID should be "new-client"


  Scenario: Update reserved lease for client
    Given I have a lease pool with client "test-client" and reserved IP "10.0.0.1"
    When I update client "test-client" ip to "10.0.0.2"
    Then Client "test-client" should have IP "10.0.0.2"

  Scenario: Update active lease for client
    Given I have a lease pool with client "test-client" and IP "10.0.0.1"
    When I update client "test-client" ip to "10.0.0.2"
    Then Client "test-client" should have IP "10.0.0.2"
