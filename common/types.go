package common

type TunnelType string

const (
	HTTP_TUN  TunnelType = "http"
	HTTPS_TUN TunnelType = "https"
	TCP_TUN   TunnelType = "tcp"
	UDP_TUN   TunnelType = "udp"
	TLS_TUN   TunnelType = "tls"
)

type InterfaceType string

const (
	SOCKS_IFACE InterfaceType = "socks"
	TCP_IFACE   InterfaceType = "tcp"
	UDP_IFACE   InterfaceType = "udp"
	TUN_IFACE   InterfaceType = "tun"
)


type PluginMode string

const (
	ENC PluginMode = "encoder"
	DEC PluginMode = "decoder"
)

type TunnelMode string

const (
	SERVER TunnelMode = "server"
	CLIENT TunnelMode = "client"
)
