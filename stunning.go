package main

import (
	"encoding/json"
	"github.com/hbahadorzadeh/stunning/common"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	socksiface "github.com/hbahadorzadeh/stunning/interface/socks"
	tcpiface "github.com/hbahadorzadeh/stunning/interface/tcp"
	tuniface "github.com/hbahadorzadeh/stunning/interface/tun"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	httptun "github.com/hbahadorzadeh/stunning/tunnel/http"
	httpstun "github.com/hbahadorzadeh/stunning/tunnel/https"
	tcptun "github.com/hbahadorzadeh/stunning/tunnel/tcp"
	tlstun "github.com/hbahadorzadeh/stunning/tunnel/tls"
	udptun "github.com/hbahadorzadeh/stunning/tunnel/udp"
	"github.com/songgao/water"
	"io/ioutil"
	"log"
	"os"
	"time"
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

func (t TunnelServer) IsAlive() bool {
	if &t.tunnelServer != nil {
		return !t.tunnelServer.Closed()
	} else {
		return false
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

func readConfig(confFile string) map[string]TunnelConfig {
	confStruct := make(map[string]TunnelConfig)
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, confStruct)
	return confStruct
}

func TunnelFactory(name string, conf TunnelConfig) TunnelCommon {
	var tun TunnelCommon
	if sorc := conf.ServiceMode; sorc != "" {
		if common.TunnelMode(sorc) == common.CLIENT {
			ttun := TunnelClient{}
			if stype := conf.ServerType; stype != "" {
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
						break
					case common.TCP_IFACE:
						ttun.interfaceClient = tcpiface.GetTcpClient(caddr, saddr, ttun.tunnelClient)
						break
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
						tuniface.GetTunIfaceClient(tconf, saddr, ttun.tunnelClient)
						break
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Conf `%s`: Invalid interface type (%s)", name, itype)
					}
				}
			}
		} else if common.TunnelMode(sorc) == common.SERVER {
			ttun := TunnelServer{}
			if stype := conf.ServerType; stype != "" {
				if saddr := conf.Listen; saddr != "" {
					switch common.TunnelType(stype) {
					case common.HTTP_TUN:
						tServer, err := httptun.StartHttpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
						break
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
						break
					case common.TCP_TUN:
						tServer, err := tcptun.StartTcpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
						break
					case common.UDP_TUN:
						tServer, err := udptun.StartUdpServer(saddr)
						if err != nil {
							log.Panicf("Conf `%s`: Failed to create tunnel server.\n%v", name, err)
						}
						ttun.tunnelServer = tServer
						break
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
						break
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
						break
					case common.TCP_IFACE:
						if iaddr := conf.Connect; iaddr != "" {
							ttun.interfaceServer = tcpiface.GetTcpServer(iaddr)
						} else {
							log.Panicf("Conf `%s`: Service connect address not specified.", name)
						}
						break
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
						break
					case common.UDP_IFACE:
					case common.SERIAL_IFACE:
					default:
						log.Panicf("Conf `%s`: Invalid interface type (%s)", name, itype)
					}
				}
			} else {
				log.Panicf("Conf `%s`: Server type not defined.", name)
			}
		} else {
			log.Panicf("Conf `%s`: Invalid service mode(%s).", name, sorc)
		}
	} else {
		log.Panicf("Conf `%s`: Service mode not specified.", name)
	}
	tun = TunnelCommon{}
	return tun
}

func main() {
	var confFile string
	for i := 1 ; i < len(os.Args); i++ {
		arg := os.Args[i]
		if len(arg) > 9 && arg[:9] == "--config=" {
			confFile = arg[9:]
		} else if len(arg) == 2 && arg[:2] == "-c" && i+1 <= len(os.Args) {
			i++
			arg := os.Args[i]
			confFile = arg
		} else if arg[:2] == "-c" && len(arg) > 2 {
			confFile = arg[2:]
		}
	}
	if confFile != "" {
		confsMap := readConfig(confFile)
		tunsMap := make(map[string]TunnelCommon)
		for name, conf := range confsMap {
			tun := TunnelFactory(name, conf)
			tunsMap[name] = tun
			tun.ListenAndServer()
		}

		for {
			for name, tun := range tunsMap {
				if !tun.IsAlive() {
					log.Printf("Tunnel `%s` is down!", name)
					go func() {
						conf, exist := confsMap[name]
						if exist {
							tun := TunnelFactory(name, conf)
							tunsMap[name] = tun
							tun.ListenAndServer()
						} else {
							log.Printf("config not found for tunnel `%s` for recreation!", name)
						}
					}()
				}
			}
			time.Sleep(time.Second)
		}
	} else {
		log.Panicf("Config file not defiend")
	}
}
