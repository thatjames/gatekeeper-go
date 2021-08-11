package dhcp

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

type Lease struct {
	ClientId string
	IP       net.IP
	Expiry   time.Time
}

type LeaseDB struct {
	start        net.IP
	leaseRange   int
	leases       map[string]*Lease
	issuedLeases map[string]*Lease
	lock         *sync.Mutex
}

func NewLeaseDB() *LeaseDB {
	return &LeaseDB{
		leases:       make(map[string]*Lease),
		issuedLeases: make(map[string]*Lease),
		lock:         new(sync.Mutex),
		start:        net.IP{10, 0, 0, 2},
	}
}

func (l *LeaseDB) GetLease(id string) (*Lease, bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	ls, ok := l.leases[id]
	return ls, ok
}

func (l *LeaseDB) AddLease(id string, ls *Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.leases[id] = ls
	l.issuedLeases[ls.IP.String()] = nil
}

func (l *LeaseDB) GetLeaseForIP(ip net.IP) (*Lease, bool) {
	lease, ok := l.issuedLeases[ip.String()]
	return lease, ok
}

func (l *LeaseDB) NextIP() net.IP {
	result := make(net.IP, 4)
	startByte := binary.BigEndian.Uint32(l.start)
	for i := startByte; i < startByte+uint32(l.leaseRange); i++ {
		binary.BigEndian.PutUint32(result, i)
		if _, ok := l.GetLeaseForIP(result); !ok {
			return result
		}
	}
	return nil
}
