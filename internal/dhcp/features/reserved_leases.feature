Feature: Reserved IP Address Management
  As a network administrator
  I want to reserve specific IP addresses for specific clients
  So that important devices always get the same IP address

  Scenario: Reserve an IP address for a client
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.10"
    When I reserve IP address "10.0.0.100" for client "test-device"
    And I request a lease for client "test-device"
    Then I should receive a lease with IP "10.0.0.100"
    And the lease state should be "reserved"
    And the lease client ID should be "test-device"
