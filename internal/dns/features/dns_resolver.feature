Feature: DNS Resolver
  In order to perform as a DNS server
  I need to be able to resolve DNS requests from clients

  Scenario: Resolve type A DNS request
    Given the DNS packet "DIaBgAABAAEAAAAABmdpdGxhYgNjb20AAAEAAcAMAAEAAQAAAP4ABKxB+04="
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr          |
      | A    | 1     | 172.65.251.78 |

  Scenario: Resolve blocked domain A record
    Given the DNS packet "pF0BAAABAAAAAAAABHRlc3QDY29tAAABAAE="
    And the domain "test.com" is blocked
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr    |
      | A    | 1     | 0.0.0.0 |

  Scenario: Resolve blocked domain AAAA record
    Given the DNS packet "ej8BAAABAAAAAAAABHRlc3QDY29tAAAcAAE="
    And the domain "test.com" is blocked
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr |
      | AAAA | 1     | ::   |

  Scenario: Resolve local domain A record
    Given the DNS packet "pF0BAAABAAAAAAAABHRlc3QDY29tAAABAAE="
    And the resolver has local domain for "test.com" with IP "172.65.251.78"
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr          |
      | A    | 1     | 172.65.251.78 |

  Scenario: Resolve non-existent A record
    Given the DNS packet "sQQBAAABAAAAAAAADm5vbnNlbnNlZG9tYWluA2NvbQAAAQAB"
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a NXDOMAIN error

  Scenario: Resolve local domain PTR record
    Given the DNS packet "6KkBAAABAAAAAAAAAjE1ATEDMTY4AzE5Mgdpbi1hZGRyBGFycGEAAAwAAQ=="
    And the resolver has local domain for "test.com" with IP "192.168.1.15"
    When I parse the DNS packet
    Then The packet should parse
    Then I resolve the DNS request
    Then I should receive a DNS response with the following answers
      | type | class | addr     |
      | PTR  | 1     | test.com |
