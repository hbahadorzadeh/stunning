package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hbahadorzadeh/stunning/core"
)

func readConfig(confFile string) map[string]core.TunnelConfig {
	confStruct := make(map[string]core.TunnelConfig)
	data, err := os.ReadFile(confFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	if err := json.Unmarshal(data, &confStruct); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
	return confStruct
}

func main() {
	var confFile string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if len(arg) > 9 && arg[:9] == "--config=" {
			confFile = arg[9:]
		} else if len(arg) == 2 && arg[:2] == "-c" && i+1 < len(os.Args) {
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
		var tunsMutex sync.RWMutex

		for name, conf := range confsMap {
			tun := core.TunnelFactory(name, conf)
			tunsMutex.Lock()
			tunsMap[name] = tun
			tunsMutex.Unlock()
			go tun.ListenAndServer()
		}

		for {
			tunsMutex.RLock()
			mapCopy := make(map[string]core.Tunnel)
			for k, v := range tunsMap {
				mapCopy[k] = v
			}
			tunsMutex.RUnlock()

			for name, tun := range mapCopy {
				if !tun.IsAlive() {
					log.Printf("Tunnel `%s` is down!", name)
					go func(tunnelName string) {
						conf, exist := confsMap[tunnelName]
						if exist {
							tun := core.TunnelFactory(tunnelName, conf)
							tunsMutex.Lock()
							tunsMap[tunnelName] = tun
							tunsMutex.Unlock()
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
		log.Fatalf("Config file not defined")
	}
}
