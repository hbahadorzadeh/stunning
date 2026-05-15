package common

import (
	"log"
	"plugin"
	"strings"
)

type PluginChain interface {
	Call(ibuff []byte) (obuff []byte)
	AddNextChainLoop(n PluginChain)
	HasNext() bool
	GetNext() PluginChain
	Close()
}

type GoPluginChain struct {
	PluginChain
	name string
	mode PluginMode
	call func(ibuff []byte) (obuff []byte)
	next PluginChain
}

func GetGoPluginChain(name string, mode PluginMode) *GoPluginChain {
	p := &GoPluginChain{
		name: name,
		mode: mode,
	}
	plug, err := plugin.Open(name)
	if err != nil {
		log.Panic(err)
	}
	v, err := plug.Lookup("Version")
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Plugin %s(version:%s)", name, *v.(*string))
	if mode == ENC {
		fe, err := plug.Lookup("Encode")
		if err != nil {
			log.Panic(err)
		}
		p.call = fe.(func(ibuff []byte) (obuff []byte))
	} else if mode == DEC {
		fd, err := plug.Lookup("Decode")
		if err != nil {
			log.Panic(err)
		}
		p.call = fd.(func(ibuff []byte) (obuff []byte))
	}
	return p
}

func (p *GoPluginChain) AddNextChainLoop(n PluginChain) {
	p.next = n
}

func (p *GoPluginChain) HasNext() bool {
	return p.next != nil
}

func (p *GoPluginChain) GetNext() PluginChain {
	return p.next
}

func (p *GoPluginChain) Call(ibuff []byte) (obuff []byte) {
	obuff = p.call(ibuff)
	if p.HasNext() {
		obuff = p.next.Call(obuff)
	}
	return obuff
}
func (*GoPluginChain) Close() {}

// CPluginChain is unused and commented out due to dl package CGO requirements
// type CPluginChain struct {
// 	PluginChain
// 	name string
// 	mode PluginMode
// 	call func(ibuff []byte) (obuff []byte)
// 	next PluginChain
// 	lib  *dl.DL
// }
//
// func GetCPluginChain(name string, mode PluginMode) *CPluginChain {
// 	...
// }

// LuaPluginChain is unused and commented out due to gopher-lua CGO requirements
// type LuaPluginChain struct {
// 	...
// }

func PluginFactory(plugins string) (input, output PluginChain) {
	var ip PluginChain
	var op PluginChain
	for _, plugin := range strings.Split(plugins, ",") {
		var nextip PluginChain
		var nextop PluginChain
		// CPP and Lua plugins are disabled due to CGO requirements
		// if plugin[:3] == "cpp" {
		// 	nextip = GetCPluginChain(plugin[4:], DEC)
		// 	nextop = GetCPluginChain(plugin[4:], ENC)
		// } else if plugin[:3] == "lua" {
		// 	nextip = GetLuaPluginChain(plugin[4:], DEC)
		// 	nextop = GetLuaPluginChain(plugin[4:], ENC)
		// } else
		if plugin[:2] == "go" {
			nextip = GetGoPluginChain(plugin[3:], DEC)
			nextop = GetGoPluginChain(plugin[3:], ENC)
		}

		if ip != nil && op != nil {
			op.AddNextChainLoop(nextop)
			nextip.AddNextChainLoop(ip)
			ip = nextip
		} else {
			ip = nextip
			op = nextop
		}
	}
	return ip, op
}
