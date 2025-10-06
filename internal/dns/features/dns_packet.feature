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

  Scenario: Parsing an invalid DNS query packet
    Given the DNS packet "tJXns9v88fdphw=="
    When I parse the DNS packet
    Then I should receive the error "DNS packet too short for header"

  Scenario: Parsing a DNS A response packet
    Given the DNS packet "xPuBgAABAAEAAAAABmdpdGxhYgNjb20AAAEAAcAMAAEAAQAAAF4ABKxB+04="
    When I parse the DNS packet
    Then I should receive a DNS packet with the IP "172.65.251.78"

  Scenario: Parsing a DNS Response with authority section
    Given the DNS packet "EjSFgAABAAEAAgACB2V4YW1wbGUDY29tAAABAAHADAABAAEAAAEsAARduNgiwAwAAgABAAFRgAAUAWEMaWFuYS1zZXJ2ZXJzA25ldADADAACAAEAAVGAABQBYgxpYW5hLXNlcnZlcnMDbmV0AAFhDGlhbmEtc2VydmVycwNuZXQAAAEAAQABUYAABMcrhzUBYgxpYW5hLXNlcnZlcnMDbmV0AAABAAEAAVGAAATHK4c1"
    When I parse the DNS packet
    Then The packet should parse
    Then I should receive a DNS packet with the IP "93.184.216.34"

  Scenario: Parsing DNS query with EDNS OPT record
    Given the DNS packet "n8oBIAABAAAAAAABBHRlc3QDY29tAAABAAEAACkE0AAAAAAADAAKAAjkbPIZw7kl8g=="
    When I parse the DNS packet
    Then The packet should parse
    Then I should receive a DNS query for "test.com"
    And the packet should have 1 additional record
    And the additional record should be an EDNS OPT record

  Scenario: Parsing a DNS Response with no answers and an authority section
    Given the DNS packet "u82BgAABAAAAAQAABG1haWwEem9obwNjb20AABwAAcARAAYAAQAAAOAANQNuczEIem9ob2NvcnDAFghkbnNhZG1pbgZ2dGl0YW7AFni0lNkAABwgAAAHCAASdQAAAA4Q"
    When I parse the DNS packet
    Then The packet should parse
    Then I should receive a DNS packet with an authority count of 1
    And the packet should have 0 answers
    And the packet should have 1 authority
