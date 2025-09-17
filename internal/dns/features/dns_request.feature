Feature: DNS Requests
  In order to provide DNS services
  As a DNS server
  I want to be able to handle DNS requests

  Scenario: A DNS request
    Given I have a DNS server
    When I send a DNS request A for "slimjim.xyz"
    Then I should receive a valid DNS response with IP "10.0.0.1"
