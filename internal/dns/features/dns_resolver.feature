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

  Scenario: A reverse lookup checks every local domain
    Given the local domain "first.lan" with IP "10.0.0.2"
    And the local domain "second.lan" with IP "10.0.0.3"
    When I ask for PTR "3.0.0.10.in-addr.arpa" over UDP
    Then the reply has RCODE NOERROR and 1 answer
    And the first answer is PTR "second.lan"
    When I ask for PTR "2.0.0.10.in-addr.arpa" over TCP
    Then the reply has RCODE NOERROR and 1 answer
    And the first answer is PTR "first.lan"

  Scenario: A private address with no local domain has no name
    Given the local domain "first.lan" with IP "10.0.0.2"
    When I ask for PTR "9.0.0.10.in-addr.arpa" over UDP
    Then the reply has RCODE NXDOMAIN and 0 answers

  Scenario: An unmatched private address goes to the forwarder for its zone
    Given the local domain "first.lan" with IP "10.0.0.2"
    And the zone "0.0.10.in-addr.arpa" is forwarded to a server answering PTR "dhcp-client.styx"
    When I ask for PTR "9.0.0.10.in-addr.arpa" over UDP
    Then the reply has RCODE NOERROR and 1 answer
    And the first answer is PTR "dhcp-client.styx"
    And the forwarder for "0.0.10.in-addr.arpa" was asked 1 time
