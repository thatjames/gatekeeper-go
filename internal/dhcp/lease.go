package dhcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thatjames/gatekeeper-go/internal/config"
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
	default:
		return "unknown"
	}
}

const (
	LeaseAvailable LeaseState = iota
	LeaseOffered
	LeaseReserved
	LeaseActive
)

type Lease struct {
	Id       int
	ClientId string
	Hostname string
	IP       net.IP
	Expiry   time.Time
	State    LeaseState
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s: %s - %s: %s expiring at %s", l.Hostname, l.IP.String(), l.ClientId, l.State, l.Expiry.Format("15:04:05"))
}

func (l *Lease) Clear() {
	l.ClientId = ""
	l.Hostname = ""
	l.Expiry = time.Time{}
	l.State = LeaseAvailable
}

type LeasePool struct {
	start             net.IP
	end               net.IP
	leases            []*Lease
	lock              *sync.Mutex
	reservedAddresses map[string]*Lease
	changed           chan struct{}
}

func NewLeasePool(startAddr, endAddr net.IP) *LeasePool {
	leaseRange := int(binary.BigEndian.Uint32(endAddr)-binary.BigEndian.Uint32(startAddr)) + 1
	start := binary.BigEndian.Uint32(startAddr)
	leases := make([]*Lease, leaseRange)
	for i := range leases {
		leases[i] = &Lease{
			IP: make(net.IP, 4),
		}
		binary.BigEndian.PutUint32(leases[i].IP, start+uint32(i))
	}
	return &LeasePool{
		start:             startAddr,
		end:               endAddr,
		lock:              new(sync.Mutex),
		leases:            leases,
		reservedAddresses: make(map[string]*Lease),
		changed:           make(chan struct{}, 1),
	}
}

// notify tells the lease writer that the lease file is out of date. It never blocks.
func (l *LeasePool) notify() {
	select {
	case l.changed <- struct{}{}:
	default:
	}
}

// SetHostname stores the hostname a client sent with its lease.
func (l *LeasePool) SetHostname(ls *Lease, hostname string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if ls.Hostname != hostname {
		ls.Hostname = hostname
		l.notify()
	}
}

func (l *LeasePool) GetLease(clientId string) *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	if reservedLease, ok := l.reservedAddresses[clientId]; ok {
		return reservedLease
	}
	for _, lease := range l.leases {
		if strings.EqualFold(clientId, lease.ClientId) {
			if time.Now().After(lease.Expiry) {
				lease.Clear()
				return nil
			}
			return lease
		}
	}
	return nil
}

func (l *LeasePool) AcceptLease(ls *Lease, ttl time.Duration) {
	if ls.State == LeaseReserved {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if lease.ClientId == ls.ClientId {
			lease.State = LeaseActive
			lease.Expiry = time.Now().Add(ttl)
			l.notify()
			return
		}
	}
}

func (l *LeasePool) UpdateLease(clientId string, ip net.IP) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if reservedLease, ok := l.reservedAddresses[clientId]; ok {
		for _, other := range l.reservedAddresses {
			if other.ClientId != clientId && other.IP.Equal(ip) {
				return errors.New("ip already reserved")
			}
		}
		reservedLease.IP = ip
		l.evictOthers(clientId, ip)
		l.notify()
		config.Lock()
		defer config.Unlock()
		return config.UpdateConfig()
	}
	for _, lease := range l.leases {
		if lease.ClientId == clientId {
			lease.IP = ip
			l.notify()
			return nil
		}
	}

	return errors.New("lease not found")
}

func (l *LeasePool) NextAvailableLease(clientId string) *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if l.isReserved(lease.IP) {
			continue
		}
		switch lease.State {
		case LeaseAvailable:
			lease.ClientId = clientId
			lease.State = LeaseOffered
			lease.Expiry = time.Now().Add(time.Second * 60)
			return lease

		case LeaseOffered:
			if strings.EqualFold(clientId, lease.ClientId) {
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			} else if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}

		case LeaseActive:
			if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.State = LeaseOffered
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}
		}
	}
	return nil
}

func (l *LeasePool) ReserveLease(clientID string, reservedIP net.IP) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.reservedAddresses {
		if lease.ClientId != clientID && lease.IP.Equal(reservedIP) {
			log.Errorf("ip %s already reserved for client %s", reservedIP.String(), lease.ClientId)
			return errors.New("ip already reserved")
		}
	}
	for _, lease := range l.leases {
		if lease.ClientId == clientID && lease.State != LeaseAvailable {
			log.Debugf("client %s already has an active lease with IP %s", clientID, lease.IP.String())
			lease.Clear()
			break
		}
	}
	l.evictOthers(clientID, reservedIP)
	l.reservedAddresses[clientID] = &Lease{ClientId: clientID, IP: reservedIP, State: LeaseReserved}
	l.notify()
	return nil
}

// isReserved reports whether ip is reserved for any client. The caller holds the lock.
func (l *LeasePool) isReserved(ip net.IP) bool {
	for _, reserved := range l.reservedAddresses {
		if reserved.IP.Equal(ip) {
			return true
		}
	}
	return false
}

// evictOthers clears the pool lease of any client other than clientID on ip. The caller holds the lock.
func (l *LeasePool) evictOthers(clientID string, ip net.IP) {
	for _, lease := range l.leases {
		if lease.IP.Equal(ip) && lease.State != LeaseAvailable && !strings.EqualFold(lease.ClientId, clientID) {
			log.Infof("clearing lease of %s on %s, now reserved for %s", lease.ClientId, ip.String(), clientID)
			lease.Clear()
		}
	}
}

func (l *LeasePool) ActiveLeases() []Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []Lease
	for i := range l.leases {
		if l.leases[i].State == LeaseActive && time.Now().Before(l.leases[i].Expiry) {
			leases = append(leases, Lease{
				ClientId: l.leases[i].ClientId,
				Hostname: l.leases[i].Hostname,
				IP:       l.leases[i].IP,
				State:    LeaseState(l.leases[i].State),
				Expiry:   l.leases[i].Expiry,
			})
		}
	}
	return leases
}

func (l *LeasePool) ReservedLeases() []Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []Lease
	for _, lease := range l.reservedAddresses {
		leases = append(leases, Lease{
			ClientId: lease.ClientId,
			Hostname: lease.Hostname,
			IP:       lease.IP,
			State:    LeaseState(lease.State),
			Expiry:   lease.Expiry,
		})
	}

	return leases
}

// ReleaseLease handles a client's DHCPRELEASE: the pool lease is cleared, a reservation is kept.
func (l *LeasePool) ReleaseLease(relLease *Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if _, ok := l.reservedAddresses[relLease.ClientId]; ok {
		log.Debugf("client %s released, keeping its reservation", relLease.ClientId)
		return
	}
	l.clearClient(relLease.ClientId)
	l.notify()
}

// DeleteLease is the admin's delete: a reservation is removed, also from the config file.
func (l *LeasePool) DeleteLease(delLease *Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if _, ok := l.reservedAddresses[delLease.ClientId]; ok {
		delete(l.reservedAddresses, delLease.ClientId)
		l.notify()
		config.Lock()
		delete(config.Config.DHCP.ReservedAddresses, delLease.ClientId)
		config.UpdateConfig()
		config.Unlock()
		return
	}
	l.clearClient(delLease.ClientId)
	l.notify()
}

// clearClient clears the pool lease of clientID. The caller holds the lock.
func (l *LeasePool) clearClient(clientID string) {
	for _, lease := range l.leases {
		if strings.EqualFold(clientID, lease.ClientId) {
			lease.Clear()
		}
	}
}
