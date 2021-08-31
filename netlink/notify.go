package netlink

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Connection struct {
	fd          int
	notifyChan  chan Message
	messageChan chan Message
	lock        *sync.Mutex
}

func New(notifyChan chan Message) (*Connection, error) {
	log.Info("Start netlink service")
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return nil, err
	}

	log.Debug("bind to unix socket")
	if err := unix.Bind(fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: unix.RTMGRP_IPV4_ROUTE,
	}); err != nil {
		return nil, err
	}

	c := &Connection{
		fd:          fd,
		notifyChan:  notifyChan,
		messageChan: make(chan Message, 100),
		lock:        new(sync.Mutex),
	}

	log.Debug("register message listener")
	go c.read()
	return c, nil
}

func (z *Connection) SendMessage(m Message) error {
	return errors.New("not implemented yet")
}

func (z *Connection) read() {
	for {
		msg, err := readPacket(z.fd)
		if err != nil {
			log.Error("unable to read from netlink module:", err.Error())
			continue
		}
		switch msg.Type {
		case unix.RTM_NEWROUTE, unix.RTM_DELROUTE:
			z.notifyChan <- msg
		default:
			z.messageChan <- msg
		}
	}
}

func readPacket(fd int) (msg Message, err error) {
	var (
		n int
		b []byte = make([]byte, os.Getpagesize())
	)
	for {
		if n, _, err = unix.Recvfrom(fd, b, unix.MSG_PEEK); err != nil {
			return
		}

		if n < len(b) {
			break
		}

		b = make([]byte, len(b)*2)
	}

	if n, _, err = unix.Recvfrom(fd, b, 0); err != nil {
		return
	}

	var hdr Header
	buff := bytes.NewBuffer(b[:n])
	if err = binary.Read(buff, binary.LittleEndian, &hdr); err != nil {
		return
	}

	msg.Header = hdr
	msg.Data = buff.Bytes()
	return
}
