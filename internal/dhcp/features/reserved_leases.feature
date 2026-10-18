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

  Scenario: A reserved client releases and rediscovers
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.10"
    And the config reserves IP "10.0.0.5" for client "02:00:00:00:00:01"
    When I reserve IP address "10.0.0.5" for client "02:00:00:00:00:01"
    And client "02:00:00:00:00:01" releases its lease
    And I request a lease for client "02:00:00:00:00:01"
    Then I should receive a lease with IP "10.0.0.5"
    And the lease state should be "reserved"
    And the config should reserve IP "10.0.0.5" for client "02:00:00:00:00:01"

  Scenario: An admin deletes a reservation
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.10"
    And the config reserves IP "10.0.0.5" for client "02:00:00:00:00:01"
    When I reserve IP address "10.0.0.5" for client "02:00:00:00:00:01"
    And the admin deletes the lease of client "02:00:00:00:00:01"
    And I request a lease for client "02:00:00:00:00:01"
    Then I should receive no lease
    And the config should have no reservation for client "02:00:00:00:00:01"

  Scenario: The pool never offers a reserved IP in range
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.3"
    When I reserve IP address "10.0.0.2" for client "02:00:00:00:00:10"
    And I request 2 leases for different clients
    Then no client but "02:00:00:00:00:10" holds IP "10.0.0.2"
    When I request the next available lease for client "02:00:00:00:00:11"
    Then I should receive no lease

  Scenario: A new reservation clears another client's lease on that IP
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.3"
    When I request the next available lease for client "02:00:00:00:00:20"
    And I accept the lease for "1h"
    And I reserve IP address "10.0.0.1" for client "02:00:00:00:00:21"
    Then client "02:00:00:00:00:20" should have no lease
    And Client "02:00:00:00:00:21" should have IP "10.0.0.1"

  Scenario: A refused reservation keeps the client's lease
    Given I have a lease pool with IP range "10.0.0.1" to "10.0.0.3"
    When I reserve IP address "10.0.0.100" for client "02:00:00:00:00:30"
    And I request the next available lease for client "02:00:00:00:00:31"
    And I accept the lease for "1h"
    And I reserve IP address "10.0.0.100" for client "02:00:00:00:00:31"
    Then Client "02:00:00:00:00:31" should have IP "10.0.0.1"
