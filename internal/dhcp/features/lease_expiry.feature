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

