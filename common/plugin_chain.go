package common

import (
	"plugin"
	"log"
	"github.com/rainycape/dl"
	"github.com/yuin/gopher-lua"
)

type PluginMode string

const(
	ENC PluginMode = "encoder"
	DEC PluginMode = "decoder"
)

type PluginChain interface {
	Call(ibuff[]byte)(obuff []byte)
	HasNext()bool
	GetNext()PluginChain
	Close()
}


type GoPluginChain struct {
	PluginChain
	name string
	mode PluginMode
	call func(ibuff[]byte)(obuff[]byte)
	next PluginChain
}

func GetGoPlugin(name string, mode PluginMode)(p GoPluginChain){
	p = GoPluginChain{
		name:name,
		mode:mode,
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
	}else if mode == DEC{
		fd, err := plug.Lookup("Decode")
		if err != nil {
			log.Panic(err)
		}
		p.call = fd.(func(ibuff []byte) (obuff []byte))
	}
	return p
}

func(p GoPluginChain)AddNextChainLoop(n PluginChain){
	p.next = n
}

func(p GoPluginChain)HasNext()bool{
	return p.next != nil
}

func(p GoPluginChain)GetNext()PluginChain{
	return p.next
}

func(p GoPluginChain)Call(ibuff[]byte)(obuff []byte){
	obuff = p.call(ibuff)
	if p.HasNext(){
		obuff = p.next.Call(obuff)
	}
	return obuff
}
func(p GoPluginChain)Close(){}


type CPluginChain struct {
	PluginChain
	name string
	mode PluginMode
	call func(ibuff[]byte)(obuff[]byte)
	next PluginChain
	lib *dl.DL
}

func GetCPluginChain(name string, mode PluginMode)(p CPluginChain){
	p = CPluginChain{
		name:name,
		mode:mode,
	}

	lib, err := dl.Open("libc", 0)
	if err != nil {
		log.Panic(err)
	}
	p.lib = lib

	var version func()string
	if err := lib.Sym("encode", &version); err != nil {
		log.Panic(err)
	}
	log.Printf("Plugin %s(version:%s)", name, version())
	if mode == ENC {
		if err := lib.Sym("encode", &p.call); err != nil {
			log.Panic(err)
		}
	}else if mode == DEC{
		if err := lib.Sym("decode", &p.call); err != nil {
			log.Panic(err)
		}
	}
	return p
}

func(p CPluginChain)AddNextChainLoop(n PluginChain){
	p.next = n
}

func(p CPluginChain)HasNext()bool{
	return p.next != nil
}

func(p CPluginChain)GetNext()PluginChain{
	return p.next
}

func(p CPluginChain)Call(ibuff[]byte)(obuff []byte){
	obuff = p.call(ibuff)
	if p.HasNext(){
		obuff = p.next.Call(obuff)
	}
	return obuff
}
func(p CPluginChain)Close(){
	p.lib.Close()
}



type LuaPluginChain struct {
	PluginChain
	name string
	mode PluginMode
	call func(ibuff[]byte)(obuff[]byte)
	next PluginChain
	pool lStatePool
	fname string
}


func GetLuaPluginChain(name string, mode PluginMode)(p LuaPluginChain){
	p = LuaPluginChain{
		name:name,
		mode:mode,
		pool: lStatePool{saved:make([]*lua.LState, 10),},
	}
	if mode == ENC {
		p.fname = "encode"
	}else if mode == DEC {
		p.fname = "decode"
	}
	return p
}

func(p LuaPluginChain)AddNextChainLoop(n PluginChain){
	p.next = n
}

func(p LuaPluginChain)HasNext()bool{
	return p.next != nil
}

func(p LuaPluginChain)GetNext()PluginChain{
	return p.next
}

func(p LuaPluginChain)Call(ibuff[]byte)(obuff []byte){
	L := p.pool.Get()
	defer p.pool.Put(L)
	if err := L.DoFile(p.name); err != nil {
		log.Panic(err)
	}
	if err := L.CallByParam(lua.P{
		Fn: L.GetGlobal(p.fname),
		NRet: 1,
		Protect: true,
	}, lua.LString(string(ibuff))); err != nil {
		panic(err)
	}
	obuff = []byte(L.Get(-1).String()) // returned value
	if p.HasNext(){
		obuff = p.next.Call(obuff)
	}

	L.DoFile("")
	return obuff
}
func(p LuaPluginChain)Close(){
	p.pool.Shutdown()
}