package dns

import (
	"testing"
)

func TestDNSOpCodeString(t *testing.T) {
	tests := []struct {
		opCode   DNSOpCode
		expected string
	}{
		{DNSOpCodeQuery, "Query"},
		{DNSOpCodeIQuery, "IQuery"},
		{DNSOpCodeStatus, "Status"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		result := tt.opCode.String()
		if result != tt.expected {
			t.Errorf("DNSOpCode(%d).String() = %s; want %s", tt.opCode, result, tt.expected)
		}
	}
}

func TestDNSRecordString(t *testing.T) {
	record := &DNSRecord{
		ParsedName: "example.com",
		Type:       DNSTypeA,
		Class:      1,
		TTL:        300,
	}
	result := record.String()
	if result == "" {
		t.Error("expected non-empty string")
	}
}

func TestDNSHeaderSetQR(t *testing.T) {
	h := &DNSHeader{}
	h.SetQR(true)
	if h.Flags&0x8000 == 0 {
		t.Error("expected QR bit to be set")
	}
	h.SetQR(false)
	if h.Flags&0x8000 != 0 {
		t.Error("expected QR bit to be cleared")
	}
}

func TestDNSHeaderSetOpcode(t *testing.T) {
	h := &DNSHeader{}
	originalFlags := h.Flags
	h.SetOpcode(DNSOpCodeStatus)
	if h.Flags == originalFlags {
		t.Error("expected flags to change")
	}
}

func TestDNSHeaderSetAA(t *testing.T) {
	h := &DNSHeader{}
	h.SetAA(true)
	if h.Flags&0x0400 == 0 {
		t.Error("expected AA bit to be set")
	}
	h.SetAA(false)
	if h.Flags&0x0400 != 0 {
		t.Error("expected AA bit to be cleared")
	}
}

func TestDNSHeaderSetTC(t *testing.T) {
	h := &DNSHeader{}
	h.SetTC(true)
	if h.Flags&0x0200 == 0 {
		t.Error("expected TC bit to be set")
	}
	h.SetTC(false)
	if h.Flags&0x0200 != 0 {
		t.Error("expected TC bit to be cleared")
	}
}

func TestDNSHeaderSetRD(t *testing.T) {
	h := &DNSHeader{}
	h.SetRD(true)
	if h.Flags&0x0100 == 0 {
		t.Error("expected RD bit to be set")
	}
}

func TestDNSHeaderSetRA(t *testing.T) {
	h := &DNSHeader{}
	h.SetRA(true)
	if h.Flags&0x0080 == 0 {
		t.Error("expected RA bit to be set")
	}
}

func TestDNSHeaderSetZ(t *testing.T) {
	h := &DNSHeader{}
	h.SetZ(1)
	if h.Flags&0x0010 == 0 {
		t.Error("expected Z bit to be set")
	}
	h.SetZ(0)
	if h.Flags&0x0010 != 0 {
		t.Error("expected Z bit to be cleared")
	}
}

func TestDNSHeaderSetRCODE(t *testing.T) {
	h := &DNSHeader{}
	h.SetRCODE(RCODENameFailure)
	if h.Flags&0x000F != 0x0003 {
		t.Error("expected RCODE to be set")
	}
}

func TestDNSHeaderRCODE(t *testing.T) {
	h := &DNSHeader{Flags: 0x0003}
	if h.RCODE() != RCODENameFailure {
		t.Errorf("expected NXDomain, got %v", h.RCODE())
	}
}

func TestNewDnsMessage(t *testing.T) {
	msg := NewDnsMessage()
	if msg == nil {
		t.Fatal("expected non-nil DNSMessage")
	}
	if msg.Header == nil {
		t.Error("expected Header to be set")
	}
	if len(msg.Questions) != 0 {
		t.Error("expected empty Questions")
	}
}

func TestDNSTypeString(t *testing.T) {
	if DNSTypeA.String() != "A" {
		t.Errorf("expected 'A', got %s", DNSTypeA.String())
	}
	if DNSTypeAAAA.String() != "AAAA" {
		t.Errorf("expected 'AAAA', got %s", DNSTypeAAAA.String())
	}
	if DNSTypeCNAME.String() != "CNAME" {
		t.Errorf("expected 'CNAME', got %s", DNSTypeCNAME.String())
	}
	if DNSTypeMX.String() != "MX" {
		t.Errorf("expected 'MX', got %s", DNSTypeMX.String())
	}
	if DNSTypeTXT.String() != "TXT" {
		t.Errorf("expected 'TXT', got %s", DNSTypeTXT.String())
	}
	if DNSTypeSOA.String() != "SOA" {
		t.Errorf("expected 'SOA', got %s", DNSTypeSOA.String())
	}
	if DNSTypeNS.String() != "NS" {
		t.Errorf("expected 'NS', got %s", DNSTypeNS.String())
	}
	if DNSTypePTR.String() != "PTR" {
		t.Errorf("expected 'PTR', got %s", DNSTypePTR.String())
	}
	if DNSType(99).String() != "unknown" {
		t.Error("expected 'unknown' for unknown type")
	}
}
