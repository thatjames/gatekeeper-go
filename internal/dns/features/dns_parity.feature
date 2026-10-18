Feature: Pi-hole parity
  In order to take Pi-hole's place on atlas
  As the LAN's DNS server
  I answer over UDP and TCP, truncate what does not fit and forward the LAN's zones to hecate

  Scenario: The same query is answered over UDP and TCP
    Given the local domain "box.test" with IP "10.0.0.2"
    When I ask for A "box.test" over UDP
    Then the reply has RCODE NOERROR and 1 answer
    When I ask for A "box.test" over TCP
    Then the reply has RCODE NOERROR and 1 answer

  Scenario: EDNS is answered with an OPT record
    Given the local domain "box.test" with IP "10.0.0.2"
    When I ask for A "box.test" over UDP with EDNS
    Then the reply has an OPT record
    When I ask for A "box.test" over UDP
    Then the reply has no OPT record

  Scenario: A reply too big for UDP is truncated and comes in full over TCP
    Given the zone "big.test" is forwarded to a server with 100 A records over TCP only
    When I ask for A "big.test" over UDP
    Then the reply is truncated with 0 answers
    When I ask for A "big.test" over UDP with EDNS
    Then the reply is truncated with 0 answers
    When I ask for A "big.test" over TCP
    Then the reply has RCODE NOERROR and 100 answers
    And the reply is not truncated

  Scenario: A reverse lookup of a LAN address is answered by hecate
    Given the zone "1.20.10.in-addr.arpa" is forwarded to a server answering PTR "atlas.styx"
    When I ask for PTR "10.1.20.10.in-addr.arpa" over UDP
    Then the reply has RCODE NOERROR and 1 answer
    And the first answer is PTR "atlas.styx"
    And the forwarder for "1.20.10.in-addr.arpa" was asked 1 time

  Scenario: A local domain's own address is not forwarded
    Given the local domain "gatekeeper.styx" with IP "10.20.1.5"
    And the zone "1.20.10.in-addr.arpa" is forwarded to a server answering PTR "atlas.styx"
    When I ask for PTR "5.1.20.10.in-addr.arpa" over TCP
    Then the reply has RCODE NOERROR and 1 answer
    And the first answer is PTR "gatekeeper.styx"
    And the forwarder for "1.20.10.in-addr.arpa" was asked 0 times

  Scenario: A forwarder that does not answer gives SERVFAIL
    Given the zone "styx" is forwarded to a server that does not answer
    When I ask for A "atlas.styx" over UDP
    Then the reply has RCODE SERVFAIL and 0 answers
