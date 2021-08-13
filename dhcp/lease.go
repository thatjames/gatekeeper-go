package dhcp

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
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
		if lease == nil {
			break
		}
		if lease.ClientId == ls.ClientId {
			l.leases[i].State = LeaseActive
			l.leases[i].Expiry = time.Now().Add(time.Second * 86400)
			return
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

func (l *LeaseDB) PeristLeases(file string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buff := make([]byte, 1)
	var counter byte = 0
	for _, lease := range l.leases {
		if lease == nil {
			break
		}
		if lease.State == LeaseActive {
			counter++
			leaseBuff := make([]byte, 0)
			leaseBuff = append(leaseBuff, byte(len(lease.ClientId)))
			leaseBuff = append(leaseBuff, []byte(lease.ClientId)...)
			leaseBuff = append(leaseBuff, lease.IP...)
			leaseBuff = append(leaseBuff, byte(lease.State))
			buff = append(buff, leaseBuff...)
		}
	}

	if counter > 0 {
		buff[0] = counter
		_, err := f.Write(buff)
		return err
	}

	return nil
}

func (l *LeaseDB) LoadLeases(file string) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err == nil {
		defer f.Close()
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	leaseCount := int(data[0])
	data = data[1:]
	for i := 0; i < leaseCount; i++ {
		var lease = new(Lease)
		cidLen := data[0]
		lease.ClientId = string(data[1 : 1+cidLen])
		data = data[1+cidLen:]
		lease.IP = data[:4]
		data = data[4:]
		lease.State = LeaseState(data[0])
		data = data[1:]
		l.leases = append(l.leases, lease)
	}
	return nil
}
