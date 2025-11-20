Feature: DNS Resolver
  In order to perform as a DNS server
  I need to be able to resolve DNS requests from clients

  Scenario: Resolve a DNS request
    Given the DNS packet "DIaBgAABAAEAAAAABmdpdGxhYgNjb20AAAEAAcAMAAEAAQAAAP4ABKxB+04="
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr          |
      | A    | 1     | 172.65.251.78 |
