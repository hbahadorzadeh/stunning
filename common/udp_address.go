package common

import (
	"net"
	"time"
)

type UdpAddress struct {
	addr     *net.UDPAddr
	lastUsed time.Time
	timeout  time.Duration
}

func GetUdpAddress(addr *net.UDPAddr) *UdpAddress {
	ua := &UdpAddress{}
	ua.addr = addr
	ua.UpdateLastUsed()
	ua.timeout = 30 * time.Second
	return ua
}

func (ua *UdpAddress) UpdateLastUsed() {
	ua.lastUsed = time.Now()
}

func (ua *UdpAddress) SetTimeout(t time.Duration) {
	ua.timeout = t
}

func (ua *UdpAddress) IsTimedOut() bool {
	return ua.lastUsed.Add(ua.timeout).Unix() < time.Now().Unix()
}

func (ua *UdpAddress) GetAddress() *net.UDPAddr {
	return ua.addr
}

func (ua *UdpAddress) Equals(uan *UdpAddress) bool {
	return ua.addr.IP.Equal(uan.addr.IP) && ua.addr.Port == ua.addr.Port
}
