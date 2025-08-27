package dhcp

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseDHCPRequest_Success(t *testing.T) {
	dat, err := ioutil.ReadFile(".test-req")
	if err != nil {
		t.Fatal(err)
	}

	m := Message(dat)

	assert.Equal(t, OpRequest, m.OpCode())
	assert.Equal(t, byte(1), m.HType())
	assert.Equal(t, byte(6), m.HLen())
	assert.Equal(t, net.HardwareAddr([]byte{0x00, 0xd8, 0x61, 0x39, 0xb5, 0x6a}), m.CHAddr())

	options := ParseOptions(m)
	assert.NotNil(t, options)
	assert.Equal(t, []byte{0x03}, options[OptionDHCPMessageType])
	assert.Equal(t, net.IP([]byte{10, 0, 0, 40}), net.IP(options[OptionRequestedIPAddress]))
	assert.Equal(t, 7, len(options[OptionParameterRequestList]))
	t.Log(options)
}
