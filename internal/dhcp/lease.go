package dhcp

import (
	"encoding/binary"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/datasource"
)

type LeasePool struct {
	start             net.IP
	end               net.IP
	leases            []*common.Lease
	lock              *sync.Mutex
	reservedAddresses map[string]*common.Lease
}

func NewLeasePool(startAddr, endAddr net.IP) *LeasePool {
	leaseRange := int(binary.BigEndian.Uint32(endAddr)-binary.BigEndian.Uint32(startAddr)) + 1
	start := binary.BigEndian.Uint32(startAddr)
	leases := make([]*common.Lease, leaseRange)
	for i := range leases {
		leases[i] = &common.Lease{
			IP: make(net.IP, 4),
		}
		binary.BigEndian.PutUint32(leases[i].IP, start+uint32(i))
	}
	return &LeasePool{
		start:             startAddr,
		end:               endAddr,
		lock:              new(sync.Mutex),
		leases:            leases,
		reservedAddresses: make(map[string]*common.Lease),
	}
}

func (l *LeasePool) GetLease(clientId string) *common.Lease {
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

func (l *LeasePool) AcceptLease(ls *common.Lease, ttl time.Duration) {
	if ls.State == common.LeaseReserved {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if lease.ClientId == ls.ClientId {
			lease.State = common.LeaseActive
			lease.Expiry = time.Now().Add(ttl)
			return
		}
	}
}

func (l *LeasePool) NextAvailableLease(clientId string) *common.Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		switch lease.State {
		case common.LeaseAvailable:
			lease.ClientId = clientId
			lease.State = common.LeaseOffered
			lease.Expiry = time.Now().Add(time.Second * 60)
			return lease

		case common.LeaseOffered:
			if strings.EqualFold(clientId, lease.ClientId) {
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			} else if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}

		case common.LeaseActive:
			if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.State = common.LeaseOffered
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}
		}
	}
	return nil
}

func (l *LeasePool) ReserveLease(clientID string, reservedIP net.IP) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.reservedAddresses[clientID] = &common.Lease{ClientId: clientID, IP: reservedIP, State: common.LeaseReserved}
}

func (l *LeasePool) PeristLeases(file string) error {
	leases := l.ReservedLeases()
	for _, lease := range l.leases {
		leases = append(leases, *lease)
	}
	return datasource.DataSource.PersistLeases(leases)
}

func (l *LeasePool) LoadLeases(file string, ttl time.Duration) error {
	leases, err := datasource.DataSource.ListLeases()
	if err != nil {
		return err
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range leases {
		for j, createdLease := range l.leases {
			if lease.IP.Equal(createdLease.IP) {
				*l.leases[j] = lease
				log.Debug("restore lease ", l.leases[j])
			}
		}
	}
	log.Debugf("loaded %d leases", len(leases))
	return nil
}

func (l *LeasePool) ActiveLeases() []common.Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []common.Lease
	for i := range l.leases {
		if l.leases[i].State == common.LeaseActive && time.Now().Before(l.leases[i].Expiry) {
			leases = append(leases, common.Lease{
				ClientId: l.leases[i].ClientId,
				Hostname: l.leases[i].Hostname,
				IP:       l.leases[i].IP,
				State:    common.LeaseState(l.leases[i].State),
				Expiry:   l.leases[i].Expiry,
			})
		}
	}
	return leases
}

func (l *LeasePool) ReservedLeases() []common.Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []common.Lease
	for _, lease := range l.reservedAddresses {
		leases = append(leases, common.Lease{
			ClientId: lease.ClientId,
			Hostname: lease.Hostname,
			IP:       lease.IP,
			State:    common.LeaseState(lease.State),
			Expiry:   lease.Expiry,
		})
	}

	return leases
}

func (l *LeasePool) ReleaseLease(relLease *common.Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if strings.EqualFold(relLease.ClientId, lease.ClientId) {
			lease.Clear()
		}
	}
}
