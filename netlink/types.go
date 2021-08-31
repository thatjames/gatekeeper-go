package netlink

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	Length uint32
	Type   uint16
	Flags  uint16
	Pid    uint32
}

type Message struct {
	Header
	Data []byte
}

func DecodeMessge(b []byte) (*Message, error) {
	var hdr Header
	buff := bytes.NewBuffer(b)
	if err := binary.Read(buff, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	return &Message{
		Header: hdr,
		Data:   buff.Bytes(),
	}, nil
}
