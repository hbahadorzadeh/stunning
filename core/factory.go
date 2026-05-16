// Package core provides the tunnel factory and main tunnel configuration.
package core

import (
	"log"
	"os"

	"github.com/hbahadorzadeh/stunning/core/common"
	icommon "github.com/hbahadorzadeh/stunning/core/interface/common"
	"github.com/hbahadorzadeh/stunning/core/metrics"
	socksiface "github.com/hbahadorzadeh/stunning/core/interface/socks"
	tcpiface "github.com/hbahadorzadeh/stunning/core/interface/tcp"
	tuniface "github.com/hbahadorzadeh/stunning/core/interface/tun"
	tcommon "github.com/hbahadorzadeh/stunning/core/tunnel/common"
	dnstun "github.com/hbahadorzadeh/stunning/core/tunnel/dns"
	h2tun "github.com/hbahadorzadeh/stunning/core/tunnel/h2"
	httptun "github.com/hbahadorzadeh/stunning/core/tunnel/http"
	httpstun "github.com/hbahadorzadeh/stunning/core/tunnel/https"
	icmptun "github.com/hbahadorzadeh/stunning/core/tunnel/icmp"
	tcptun "github.com/hbahadorzadeh/stunning/core/tunnel/tcp"
	tlstun "github.com/hbahadorzadeh/stunning/core/tunnel/tls"
	udptun "github.com/hbahadorzadeh/stunning/core/tunnel/udp"
	udpstun "github.com/hbahadorzadeh/stunning/core/tunnel/udps"
	wstun "github.com/hbahadorzadeh/stunning/core/tunnel/ws"
	"github.com/songgao/water"
)

type TunnelConfig struct {
	Cert          string
	Connect       string
	DeviceName    string
	InterfaceType string
	Key           string
	Listen        string
	Mtu           string
	ServerType    string
	ServiceMode   string
}

type Tunnel interface {
	ListenAndServer()
	IsAlive() bool
	GetMetrics() *metrics.Metrics
}
type TunnelCommon struct {
	Tunnel
	tunnelType    common.TunnelType
	interfaceType common.InterfaceType
	tunnelMode    common.TunnelMode
	metrics       *metrics.Metrics
}

type TunnelServer struct {
	TunnelCommon
	tunnelServer    tcommon.TunnelServer
	interfaceServer icommon.TunnelInterfaceServer
}

type TunnelClient struct {
	TunnelCommon
	serverAddress   string
	tunnelClient    tcommon.TunnelDialer
	interfaceClient icommon.TunnelInterfaceClient
}

func (t TunnelCommon) GetTunnelType() common.TunnelType {
	return t.tunnelType
}

func (t TunnelCommon) GetInterfaceType() common.InterfaceType {
	return t.interfaceType
}

func (t TunnelCommon) GetTunnelMode() common.TunnelMode {
	return t.tunnelMode
}

func (t TunnelCommon) GetMetrics() *metrics.Metrics {
	return t.metrics
}

func (t TunnelServer) ListenAndServer() {
	if t.tunnelServer != nil {
		defer t.tunnelServer.Close()
		t.tunnelServer.WaitingForConnection()
	} else {
		log.Panic("No tunnel server")
	}
}

func (t TunnelServer) IsAlive() bool {
	if t.tunnelServer != nil {
		return !t.tunnelServer.Closed()
	}
	return false
}

func (t TunnelClient) ListenAndServer() {
	if t.interfaceClient != nil {
		defer t.interfaceClient.Close()
		t.tunnelClient.Dial("", t.serverAddress)
	} else {
		log.Panic("No tunnel Interface")
	}
}

func (t TunnelClient) IsAlive() bool {
	if t.interfaceClient != nil {
		return !t.interfaceClient.Closed()
	}
	return false
}

func TunnelFactory(name string, conf TunnelConfig) Tunnel {
	if sorc := conf.ServiceMode; sorc != "" {
		if common.TunnelMode(sorc) == common.CLIENT {
			ttun := TunnelClient{
				TunnelCommon: TunnelCommon{
					metrics: metrics.NewMetrics(),
				},
			}
			if stype := conf.ServerType; stype != "" {
				switch common.TunnelType(stype) {
				case common.HTTP_TUN:
					ttun.tunnelClient = httptun.GetHttpDialer()
				case common.HTTPS_TUN:
					ttun.tunnelClient = httpstun.GetHttpsDialer()
				case common.TCP_TUN:
					ttun.tunnelClient = tcptun.GetTcpDialer()
				case common.UDP_TUN:
					ttun.tunnelClient = udptun.GetUdpDialer()
				case common.TLS_TUN:
					ttun.tunnelClient = tlstun.GetTlsDialer()
				case common.H2_TUN:
					ttun.tunnelClient = h2tun.GetH2Dialer()
				case common.WS_TUN:
					ttun.tunnelClient = wstun.GetWsDialer()
				case common.UDPS_TUN:
					ttun.tunnelClient = udpstun.GetUdpsDialer()
				case common.DNS_TUN:
					ttun.tunnelClient = dnstun.GetDnsDialer()
				case common.ICMP_TUN:
					ttun.tunnelClient = icmptun.GetIcmpDialer()
				default:
					log.Panicf("Conf `%s`: Invalid server type(%s).", name, stype)
				}
				saddr := conf.Connect
				if saddr == "" {
					log.Panicf("Conf `%s`: Service connect address not specified.", name)
				}
				caddr := conf.Listen
				if caddr == "" {
					log.Panicf("Conf `%s`: Service listen address not specified.", name)
				}

				if itype := conf.InterfaceType; itype != "" {
					switch common.InterfaceType(itype) {
					case common.SOCKS_IFACE:
						ttun.interfaceClient = socksiface.GetSocksClient(caddr, saddr, ttun.tunnelClient)
					case common.TCP_IFACE:
						ttun.interfaceClient = tcpiface.GetTcpClient(caddr, saddr, ttun.tunnelClient)
					case common.TUN_IFACE:
						imtu := conf.Mtu
						if imtu == "" {
							imtu = "1500"
						}
						iname := conf.DeviceName
						if iname == "" {
							iname = "tun"
						}
						tconf := tuniface.TunConfig{
							DevType: water.TUN,
							Address: caddr,
							Name:    iname,
							MTU:     imtu,
						}
						ttun.interfaceClient = tuniface.GetTunIfaceClient(tconf, saddr, ttun.tunnelClient)
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Conf `%s`: Invalid interface type (%s)", name, itype)
					}
				}
			}
			metrics.GetGlobalCollector().Register(name, ttun.GetMetrics())
			return &ttun
		} else if common.TunnelMode(sorc) == common.SERVER {
			ttun := TunnelServer{
				TunnelCommon: TunnelCommon{
					metrics: metrics.NewMetrics(),
				},
			}
			if stype := conf.ServerType; stype != "" {
				if saddr := conf.Listen; saddr != "" {
					switch common.TunnelType(stype) {
					case common.HTTP_TUN:
						tServer, err := httptun.StartHttpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.HTTPS_TUN:
						scert := conf.Cert
						skey := conf.Key
						if scert == "" {
							log.Panicf("Conf `%s`: Cert not defiend", name)
						} else if _, err := os.Stat(scert); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Cert file not exist", name)
						}
						if skey == "" {
							log.Panicf("Conf `%s`: Key not defiend", name)
						} else if _, err := os.Stat(skey); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Key file not exist", name)
						}

						tServer, err := httpstun.StartHttpsServer(scert, skey, saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.TCP_TUN:
						tServer, err := tcptun.StartTcpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.UDP_TUN:
						tServer, err := udptun.StartUdpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.TLS_TUN:
						scert := conf.Cert
						skey := conf.Key
						if scert == "" {
							log.Panicf("Conf `%s`: Cert not defiend", name)
						} else if _, err := os.Stat(scert); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Cert file not exist", name)
						}
						if skey == "" {
							log.Panicf("Conf `%s`: Key not defiend", name)
						} else if _, err := os.Stat(skey); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Key file not exist", name)
						}

						tServer, err := tlstun.StartTlsServer(scert, skey, saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.H2_TUN:
						scert := conf.Cert
						skey := conf.Key
						if scert == "" {
							log.Panicf("Conf `%s`: Cert not defiend", name)
						} else if _, err := os.Stat(scert); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Cert file not exist", name)
						}
						if skey == "" {
							log.Panicf("Conf `%s`: Key not defiend", name)
						} else if _, err := os.Stat(skey); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Key file not exist", name)
						}
						tServer, err := h2tun.StartH2Server(scert, skey, saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.WS_TUN:
						scert := conf.Cert
						skey := conf.Key
						if scert == "" {
							log.Panicf("Conf `%s`: Cert not defiend", name)
						} else if _, err := os.Stat(scert); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Cert file not exist", name)
						}
						if skey == "" {
							log.Panicf("Conf `%s`: Key not defiend", name)
						} else if _, err := os.Stat(skey); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Key file not exist", name)
						}
						tServer, err := wstun.StartWsServer(scert, skey, saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.UDPS_TUN:
						scert := conf.Cert
						skey := conf.Key
						if scert == "" {
							log.Panicf("Conf `%s`: Cert not defiend", name)
						} else if _, err := os.Stat(scert); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Cert file not exist", name)
						}
						if skey == "" {
							log.Panicf("Conf `%s`: Key not defiend", name)
						} else if _, err := os.Stat(skey); os.IsNotExist(err) {
							log.Panicf("Conf `%s`: Key file not exist", name)
						}
						tServer, err := udpstun.StartUdpsServer(scert, skey, saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.DNS_TUN:
						tServer, err := dnstun.StartDnsServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					case common.ICMP_TUN:
						log.Printf("WARNING: ICMP tunnel requires root or CAP_NET_RAW privileges")
						tServer, err := icmptun.StartIcmpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
					default:
						log.Panicf("Conf `%s`: Invalid server type(%s).", name, stype)
					}
				} else {
					log.Panicf("Conf `%s`: Service listen address not specified.", name)
				}

				if itype := conf.InterfaceType; itype != "" {
					switch common.InterfaceType(itype) {
					case common.SOCKS_IFACE:
						ttun.interfaceServer = socksiface.GetSocksServer()
					case common.TCP_IFACE:
						if iaddr := conf.Connect; iaddr != "" {
							ttun.interfaceServer = tcpiface.GetTcpServer(iaddr)
						} else {
							log.Panicf("Conf `%s`: Service connect address not specified.", name)
						}
					case common.TUN_IFACE:
						iaddr := conf.Connect
						if iaddr == "" {
							log.Panicf("Conf `%s`: Service connect address not specified.", name)
						}
						imtu := conf.Mtu
						if imtu == "" {
							imtu = "1500"
						}
						iname := conf.DeviceName
						if iname == "" {
							iname = "tun"
						}
						tconf := tuniface.TunConfig{
							DevType: water.TUN,
							Address: iaddr,
							Name:    iname,
							MTU:     imtu,
						}
						ttun.interfaceServer = tuniface.GetTunIface(tconf)
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Conf `%s`: Invalid interface type (%s)", name, itype)
					}
				}
			} else {
				log.Panicf("Conf `%s`: Server type not defined.", name)
			}
			metrics.GetGlobalCollector().Register(name, ttun.GetMetrics())
			return &ttun
		}
		log.Panicf("Conf `%s`: Invalid service mode(%s).", name, sorc)
	}
	log.Panicf("Conf `%s`: Service mode not specified.", name)
	return nil
}
