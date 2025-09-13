Feature: Lease Pool Initialization
  As a DHCP server
  I want to initialize a pool of available IP addresses
  So that I can assign them to requesting clients

  Scenario: Initialize lease pool with IP range
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.95"
    Then all leases should initially be available

  Scenario: Assign all available leases
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.10"
    When I request 10 leases for different clients
    Then the lease pool should contain 10 active leases

