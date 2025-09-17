Feature: DNS Packet Parsing
  In order to provide DNS services
  As a DNS server
  I want to be able to parse DNS packets
  
  Scenario: Parsing DNS A query packet
    Given the DNS packet "xPsBAAABAAAAAAAABmdpdGxhYgNjb20AAAEAAQ=="
    When I parse the DNS packet
    Then I should receive a DNS query for "gitlab.com"
    
  Scenario: Parsing DNS AAAA query packet  
    Given the DNS packet "//QBAAABAAAAAAAABmdpdGxhYgNjb20AABwAAQ=="
    When I parse the DNS packet
    Then I should receive a DNS query for "gitlab.com"
