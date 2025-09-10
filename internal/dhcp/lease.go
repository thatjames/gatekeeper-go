package dhcp

import (
	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
)

type LeasePool struct {
	start             net.IP
	end               net.IP
	leases            []*common.Lease
	lock              *sync.Mutex
	reservedAddresses map[string]*common.Lease
}

func NewLeasePool(startAddr, endAddr net.IP) *LeasePool {
	leaseRange := int(binary.BigEndian.Uint32(endAddr) - binary.BigEndian.Uint32(startAddr))
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
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buff := make([]byte, 1)
	var counter byte = 0
	l.lock.Lock()
	for _, lease := range l.leases {
		if lease == nil {
			break
		}
		if lease.State == common.LeaseActive {
			counter++
			leaseBuff := make([]byte, 0)
			leaseBuff = append(leaseBuff, byte(len(lease.ClientId)))
			leaseBuff = append(leaseBuff, []byte(lease.ClientId)...)
			leaseBuff = append(leaseBuff, byte(len(lease.Hostname)))
			leaseBuff = append(leaseBuff, []byte(lease.Hostname)...)
			leaseBuff = append(leaseBuff, lease.IP...)
			leaseBuff = append(leaseBuff, byte(lease.State))
			buff = append(buff, leaseBuff...)
		}
	}
	l.lock.Unlock()

	if counter > 0 {
		buff[0] = counter
		_, err := f.Write(buff)
		return err
	}

	return nil
}

func (l *LeasePool) LoadLeases(file string, ttl time.Duration) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	switch {
	case err == nil:
		defer f.Close()
	case os.IsNotExist(err):
		return nil
	default:
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}
	leaseCount := int(data[0])
	data = data[1:]
	leases := make([]*common.Lease, 0)
	for i := 0; i < leaseCount; i++ {
		var lease = new(common.Lease)
		cidLen := data[0]
		lease.ClientId = string(data[1 : 1+cidLen])
		data = data[1+cidLen:]
		hostLen := data[0]
		lease.Hostname = string(data[1 : 1+hostLen])
		data = data[1+hostLen:]
		lease.IP = data[:4]
		data = data[4:]
		lease.State = common.LeaseState(data[0])
		lease.Expiry = time.Now().Add(ttl)
		data = data[1:]
		leases = append(leases, lease)
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range leases {
		for j, createdLease := range l.leases {
			if lease.IP.Equal(createdLease.IP) {
				*l.leases[j] = *lease
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
