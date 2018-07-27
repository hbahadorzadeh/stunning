package stunning

import (
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	tlstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tls"
	tcptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/tcp"
	udptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/udp"
	httptun "gitlab.com/h.bahadorzadeh/stunning/tunnel/http"
	httpstun "gitlab.com/h.bahadorzadeh/stunning/tunnel/https"
	"gitlab.com/h.bahadorzadeh/stunning/common"
	"log"
)

type Tunnel struct {
	tunnelType common.TunnelType
	interfaceType common.InterfaceType
	tunnelMode common.TunnelMode
	inputPluginChain common.PluginChain
	outputPluginChain common.PluginChain
	tunnelClient tcommon.TunnelDialer
	tunnelServer tcommon.TunnelServer
	interfaceClient icommon.TunnelInterfaceClient
	interfaceServer icommon.TunnelInterfaceServer
}

func(t Tunnel)GetTunnelType()common.TunnelType{
	return t.tunnelType
}

func(t Tunnel)GetInterfaceType()common.InterfaceType{
	return t.interfaceType
}

func(t Tunnel)GetTunnelMode()common.TunnelMode{
	return t.tunnelMode
}


func(t Tunnel)ListenAndServer(){
	if t.tunnelMode == common.SERVER {
		if &t.tunnelServer != nil{
			defer t.tunnelServer.Close()
			t.tunnelServer.WaitingForConnection()
		}else{
			log.Panic("No tunnel server")
		}
	}else if t.tunnelMode == common.CLIENT {
		if &t.interfaceClient != nil{
			defer t.interfaceClient.Close()
			t.tunnelServer.WaitingForConnection()
		}else{
			log.Panic("No tunnel Interface")
		}
	}
}

func TunnelFactory(conf map[string]string)Tunnel{
	tun := Tunnel{

	}
}