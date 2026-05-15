package udps

import (
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"github.com/pion/dtls/v2"
	"net"
)

type UdpsDialer struct {
	tcommon.TunnelDialer
}

func GetUdpsDialer() UdpsDialer {
	return UdpsDialer{}
}

func (d UdpsDialer) Dial(network, addr string) (c net.Conn, err error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	config := &dtls.Config{
		InsecureSkipVerify: false,
	}

	conn, err := dtls.Dial("udp", udpAddr, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (d UdpsDialer) Protocol() tcommon.TunnelProtocol {
	return tcommon.Udps
}
