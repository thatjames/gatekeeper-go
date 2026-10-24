package dhcp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// syntheticRequest builds a DHCPREQUEST from a locally administered MAC
// asking for 10.0.0.40 with a seven-entry parameter request list.
func syntheticRequest() []byte {
	pkt := make([]byte, 300)
	pkt[0], pkt[1], pkt[2] = 1, 1, 6               // op BOOTREQUEST, htype ethernet, hlen 6
	copy(pkt[4:8], []byte{0x8d, 0x18, 0xab, 0x3f}) // xid
	copy(pkt[28:34], []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01})
	copy(pkt[236:240], []byte{0x63, 0x82, 0x53, 0x63}) // magic cookie
	copy(pkt[240:], []byte{
		53, 1, 3, // DHCP message type: REQUEST
		50, 4, 10, 0, 0, 40, // requested IP address
		55, 7, 1, 28, 2, 3, 15, 6, 12, // parameter request list
		255, // end
	})
	return pkt
}

func Test_ParseDHCPRequest_Success(t *testing.T) {
	m := Message(syntheticRequest())

	assert.Equal(t, OpRequest, m.OpCode())
	assert.Equal(t, byte(1), m.HType())
	assert.Equal(t, byte(6), m.HLen())
	assert.Equal(t, net.HardwareAddr([]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01}), m.CHAddr())

	options := ParseOptions(m)
	assert.NotNil(t, options)
	assert.Equal(t, []byte{0x03}, options[OptionDHCPMessageType])
	assert.Equal(t, net.IP([]byte{10, 0, 0, 40}), net.IP(options[OptionRequestedIPAddress]))
	assert.Equal(t, 7, len(options[OptionParameterRequestList]))
}
