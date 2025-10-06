Feature: Lease Expiry Management
  As a DHCP server
  I want to handle expired leases appropriately
  So that IP addresses can be reclaimed and reused

  Scenario: Get expired lease returns nothing
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.95"
    When I request the next available lease for client "test-client"
    And the lease for client "test-client" expires in the past
    And I request a lease for client "test-client"
    Then I should receive no lease

  Scenario: Expired lease is given to a new client
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.3"
    And there is an expired lease for client "old-client" with IP "10.0.0.1"
    When I request the first expired lease for client "test-client"
    Then I should receive a lease
    And the lease client ID should be "test-client"
    And the IP address should be "10.0.0.1"
