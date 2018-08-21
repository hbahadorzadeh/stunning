package stunning

import (
	"github.com/songgao/water"
	"gitlab.com/h.bahadorzadeh/stunning/common"
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	socksiface "gitlab.com/h.bahadorzadeh/stunning/interface/socks"
	tcpiface "gitlab.com/h.bahadorzadeh/stunning/interface/tcp"
	tuniface "gitlab.com/h.bahadorzadeh/stunning/interface/tun"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	httptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/http"
	httpstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/https"
	tcptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tcp"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	udptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/udp"
	"log"
)

type Tunnel interface {
	ListenAndServer()
}
type TunnelCommon struct {
	Tunnel
	tunnelType        common.TunnelType
	interfaceType     common.InterfaceType
	tunnelMode        common.TunnelMode
	inputPluginChain  common.PluginChain
	outputPluginChain common.PluginChain
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

func (t TunnelServer) ListenAndServer() {
	if &t.tunnelServer != nil {
		defer t.tunnelServer.Close()
		t.tunnelServer.WaitingForConnection()
	} else {
		log.Panic("No tunnel server")
	}
}

func (t TunnelClient) ListenAndServer() {
	if &t.interfaceClient != nil {
		defer t.interfaceClient.Close()
		t.tunnelClient.Dial("", t.serverAddress)
	} else {
		log.Panic("No tunnel Interface")
	}
}

func TunnelFactory(conf map[string]string) TunnelCommon {
	var tun TunnelCommon
	if sorc, exist := conf["service_mode"]; exist {
		if common.TunnelMode(sorc) == common.CLIENT {
			ttun := TunnelClient{}
			if stype, exist := conf["server_type"]; exist {
				switch common.TunnelType(stype) {
				case common.HTTP_TUN:
					ttun.tunnelClient = httptun.GetHttpDialer()
					break
				case common.HTTPS_TUN:
					ttun.tunnelClient = httpstun.GetHttpsDialer()
					break
				case common.TCP_TUN:
					ttun.tunnelClient = tcptun.GetTcpDialer()
					break
				case common.UDP_TUN:
					ttun.tunnelClient = udptun.GetUdpDialer()
					break
				case common.TLS_TUN:
					ttun.tunnelClient = tlstun.GetTlsDialer()
					break
				default:
					log.Panicf("Invalid server type(%s).", stype)
				}
				saddr, sexist := conf["connect"]
				if !sexist {
					log.Panicf("Service connect address not specified.")
				}
				caddr, cexist := conf["listen"]
				if !cexist {
					log.Panicf("Service listen address not specified.")
				}

				if itype, exist := conf["interface_type"]; exist {
					switch common.InterfaceType(itype) {
					case common.SOCKS_IFACE:
						ttun.interfaceClient = socksiface.GetSocksClient(caddr, saddr, ttun.tunnelClient)
						break
					case common.TCP_IFACE:
						ttun.interfaceClient = tcpiface.GetTcpClient(caddr, saddr, ttun.tunnelClient)
						break
					case common.TUN_IFACE:
						imtu, exist := conf["mtu"]
						if !exist {
							imtu = "1500"
						}
						iname, exist := conf["devname"]
						if !exist {
							iname = "tun"
						}
						conf := tuniface.TunConfig{
							DevType: water.TUN,
							Address: caddr,
							Name:    iname,
							MTU:     imtu,
						}
						tuniface.GetTunIfaceClient(conf, saddr, ttun.tunnelClient)
						break
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Invalid interface type (%s)", itype)
					}
				}
			}
		} else if common.TunnelMode(sorc) == common.SERVER {
			ttun := TunnelServer{}
			if stype, exist := conf["server_type"]; exist {
				if saddr, exist := conf["listen"]; exist {
					switch common.TunnelType(stype) {
					case common.HTTP_TUN:
						tServer, err := httptun.StartHttpServer(saddr)
						if err != nil {
							log.Panicf("Failed to create tunnel server.\n%v", err)
						}
						ttun.tunnelServer = tServer
						break
					case common.HTTPS_TUN:
						scert, cexist := conf["cert"]
						skey, kexist := conf["key"]
						if cexist && kexist {
							tServer, err := httpstun.StartHttpsServer(scert, skey, saddr)
							if err != nil {
								log.Panicf("Failed to create tunnel server.\n%v", err)
							}
							ttun.tunnelServer = tServer
						} else {
							log.Panicf("Key or Cert not defiend")
						}
						break
					case common.TCP_TUN:
						tServer, err := tcptun.StartTcpServer(saddr)
						if err != nil {
							log.Panicf("Failed to create tunnel server.\n%v", err)
						}
						ttun.tunnelServer = tServer
						break
					case common.UDP_TUN:
						tServer, err := udptun.StartUdpServer(saddr)
						if err != nil {
							log.Panicf("Failed to create tunnel server.\n%v", err)
						}
						ttun.tunnelServer = tServer
						break
					case common.TLS_TUN:
						scert, cexist := conf["cert"]
						skey, kexist := conf["key"]
						if cexist && kexist {
							tServer, err := tlstun.StartTlsServer(scert, skey, saddr)
							if err != nil {
								log.Panicf("Failed to create tunnel server.\n%v", err)
							}
							ttun.tunnelServer = tServer
						} else {
							log.Panicf("Key or Cert not defiend")
						}
						break
					default:
						log.Panicf("Invalid server type(%s).", stype)
					}
				} else {
					log.Panicf("Service listen address not specified.")
				}

				if itype, exist := conf["interface_type"]; exist {
					switch common.InterfaceType(itype) {
					case common.SOCKS_IFACE:
						ttun.interfaceServer = socksiface.GetSocksServer()
						break
					case common.TCP_IFACE:
						if iaddr, exist := conf["connect"]; exist {
							ttun.interfaceServer = tcpiface.GetTcpServer(iaddr)
						} else {
							log.Panicf("Service connect address not specified.")
						}
						break
					case common.TUN_IFACE:
						iaddr, exist := conf["connect"]
						if !exist {
							log.Panicf("Service connect address not specified.")
						}
						imtu, exist := conf["mtu"]
						if !exist {
							imtu = "1500"
						}
						iname, exist := conf["devname"]
						if !exist {
							iname = "tun"
						}
						conf := tuniface.TunConfig{
							DevType: water.TUN,
							Address: iaddr,
							Name:    iname,
							MTU:     imtu,
						}
						ttun.interfaceServer = tuniface.GetTunIface(conf)
						break
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Invalid interface type (%s)", itype)
					}
				}
			} else {
				log.Panicf("Server type not defined.")
			}
		} else {
			log.Panicf("Invalid service mode(%s).", sorc)
		}
	} else {
		log.Panicf("Service mode not specified.")
	}
	tun = TunnelCommon{}
	return tun
}
