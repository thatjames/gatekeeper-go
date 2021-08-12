package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type LeaseState int

func (ls LeaseState) String() string {
	switch ls {
	case LeaseAvailable:
		return "Available"
	case LeaseOffered:
		return "Offered"
	case LeaseReserved:
		return "Reserved"
	case LeaseActive:
		return "Active"
	case LeaseExpired:
		return "Expired"
	default:
		return "unknown"
	}
}

const (
	LeaseAvailable LeaseState = iota + 1
	LeaseOffered
	LeaseReserved
	LeaseActive
	LeaseExpired
)

type Lease struct {
	ClientId string
	IP       net.IP
	Expiry   time.Time
	State    LeaseState
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s - %s: %s expiring at %s", l.IP.String(), l.ClientId, l.State, l.Expiry.Format("15:04:05"))
}

type LeaseDB struct {
	start  net.IP
	end    net.IP
	leases []*Lease
	lock   *sync.Mutex
}

func NewLeaseDB(startAddr, endAddr net.IP) *LeaseDB {
	leaseRange := int(binary.BigEndian.Uint32(endAddr) - binary.BigEndian.Uint32(startAddr))
	return &LeaseDB{
		start:  startAddr,
		end:    endAddr,
		lock:   new(sync.Mutex),
		leases: make([]*Lease, leaseRange),
	}
}

func (l *LeaseDB) GetLease(clientId string) *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if lease != nil && strings.EqualFold(clientId, lease.ClientId) {
			return lease
		}
	}
	return nil
}

func (l *LeaseDB) AcceptLease(ls *Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for i, lease := range l.leases {
		if lease.ClientId == ls.ClientId {
			l.leases[i].State = LeaseActive
		}
	}
}

func (l *LeaseDB) NextAvailableLease(clientId string) *Lease {
	start := binary.BigEndian.Uint32(l.start)
	l.lock.Lock()
	defer l.lock.Unlock()
	for i, lease := range l.leases {
		if lease == nil { //empty slot
			l.leases[i] = &Lease{
				ClientId: clientId,
				State:    LeaseOffered,
				IP:       make(net.IP, 4),
			}
			binary.BigEndian.PutUint32(l.leases[i].IP, start+uint32(i))
			return l.leases[i]
		}
	}
	return nil
}

func (l *LeaseDB) ReserveLease(clientID string, reservedIP net.IP) {
	l.lock.Lock()
	defer l.lock.Unlock()
	//Reserved leases go to the top
	leases := make([]*Lease, 0, len(l.leases)+1)
	leases = append(leases, &Lease{ClientId: clientID, IP: reservedIP, State: LeaseReserved})
	l.leases = append(leases, l.leases...)
}
