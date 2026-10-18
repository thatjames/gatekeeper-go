Feature: DNS Requests
  In order to provide DNS services
  As a DNS server
  I want to be able to handle DNS requests

  Scenario: A DNS request
    Given the resolver has a cache for "slimjim.xyz" with IP "10.0.0.1"
    When I send a DNS request A for "slimjim.xyz"
    Then I should receive a single valid DNS response with IP "10.0.0.1"

  Scenario: A query without exactly one question gets FORMERR and the server keeps answering
    Given the local domain "box.test" with IP "10.0.0.2"
    When I send a query with 0 questions over UDP
    Then the reply has RCODE FORMERR and 0 answers
    When I send a query with 2 questions over UDP
    Then the reply has RCODE FORMERR and 0 answers
    When I send a query with 0 questions over TCP
    Then the reply has RCODE FORMERR and 0 answers
    When I ask for A "box.test" over UDP
    Then the reply has RCODE NOERROR and 1 answer
