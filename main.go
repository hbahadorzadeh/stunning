package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/hbahadorzadeh/stunning/core"
)

func readConfig(confFile string) map[string]core.TunnelConfig {
	confStruct := make(map[string]core.TunnelConfig)
	data, err := os.ReadFile(confFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &confStruct); err != nil {
		panic(err)
	}
	return confStruct
}

func main() {
	var confFile string
	for i := 1; i < len(os.Args); i++ {
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
		tunsMap := make(map[string]core.Tunnel)
		for name, conf := range confsMap {
			tun := core.TunnelFactory(name, conf)
			tunsMap[name] = tun
			go tun.ListenAndServer()
		}

		for {
			for name, tun := range tunsMap {
				if !tun.IsAlive() {
					log.Printf("Tunnel `%s` is down!", name)
					go func(tunnelName string) {
						conf, exist := confsMap[tunnelName]
						if exist {
							tun := core.TunnelFactory(tunnelName, conf)
							tunsMap[tunnelName] = tun
							go tun.ListenAndServer()
						} else {
							log.Printf("config not found for tunnel `%s` for recreation!", tunnelName)
						}
					}(name)
				}
			}
			time.Sleep(time.Second)
		}
	} else {
		log.Panicf("Config file not defiend")
	}
}
