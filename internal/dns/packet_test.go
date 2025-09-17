package dns

import (
	"encoding/base64"
	"testing"
)

func Test_ParseDNSARecordRequest_GitLabDotCom(t *testing.T) {
	// This is a DNS QUERY for A record of gitlab.com
	testPacket, err := base64.StdEncoding.DecodeString("xPsBAAABAAAAAAAABmdpdGxhYgNjb20AAAEAAQ==")
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	if len(testPacket) < 12 {
		t.Fatalf("Packet too short: %d bytes", len(testPacket))
	}

	record, err := ParseDNSPacket(testPacket)
	if err != nil {
		t.Fatalf("Failed to parse DNS packet: %v", err)
	}

	// Test header
	if record.Header.QR != 0 {
		t.Errorf("Expected query (QR=0), got QR=%d", record.Header.QR)
	}

	if record.Header.Opcode != DNSOpCodeQuery {
		t.Errorf("Expected standard query, got opcode %v", record.Header.Opcode)
	}

	if record.Header.RD != 1 {
		t.Errorf("Expected recursion desired, got RD=%d", record.Header.RD)
	}

	// Test questions
	if len(record.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(record.Questions))
	}

	question := record.Questions[0]
	if question.Name != "gitlab.com" {
		t.Errorf("Expected name 'gitlab.com', got '%s'", question.Name)
	}

	if question.Type != DNSTypeA {
		t.Errorf("Expected A record type, got %v", question.Type)
	}

	if question.Class != 1 {
		t.Errorf("Expected class 1 (IN), got %d", question.Class)
	}

	// This is a query - should have NO answers
	if len(record.Answers) != 0 {
		t.Errorf("Expected 0 answers in query, got %d", len(record.Answers))
	}

	if len(record.Authorities) != 0 {
		t.Errorf("Expected 0 authorities in query, got %d", len(record.Authorities))
	}

	if len(record.Additionals) != 0 {
		t.Errorf("Expected 0 additionals in query, got %d", len(record.Additionals))
	}
}

func Test_ParseDNSAAAARecordRequest_GitLabDotCom(t *testing.T) {
	testPacket, err := base64.StdEncoding.DecodeString("//QBAAABAAAAAAAABmdpdGxhYgNjb20AABwAAQ==")
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	record, err := ParseDNSPacket(testPacket)
	if err != nil {
		t.Fatalf("Failed to parse DNS packet: %v", err)
	}

	// Verify this is an AAAA query
	if len(record.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(record.Questions))
	}

	question := record.Questions[0]
	if question.Name != "gitlab.com" {
		t.Errorf("Expected name 'gitlab.com', got '%s'", question.Name)
	}

	if question.Type != DNSTypeAAAA {
		t.Errorf("Expected AAAA record type, got %v", question.Type)
	}

	// Query should have no answers
	if len(record.Answers) != 0 {
		t.Errorf("Expected 0 answers in query, got %d", len(record.Answers))
	}
}
