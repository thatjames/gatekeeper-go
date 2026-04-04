package util

import (
	"testing"
)

func TestByteSize(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1K"},
		{1536, "1.5K"},
		{1048576, "1M"},
		{1572864, "1.5M"},
		{1073741824, "1G"},
		{1610612736, "1.5G"},
	}

	for _, tt := range tests {
		result := ByteSize(tt.input)
		if result != tt.expected {
			t.Errorf("ByteSize(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}
