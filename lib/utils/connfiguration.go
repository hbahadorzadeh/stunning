package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type StunningConfig struct {
	*viper.Viper
	tunConf map[string]*TunnelConfig
}

type ApplicationConfig struct {
	LogLevel    int
	LogLocation string
	Crt         string
	Key         string
}

type TunnelConfig struct {
	TunnelType  TunnelType
	IsClient    bool
	ListenAddr  string
	ConnectAddr string
}

type TunnelType int

const (
	Nil   TunnelType = -1
	Tcp   TunnelType = 0
	Udp   TunnelType = 1
	Socks TunnelType = 2
	Tun   TunnelType = 3
	Tap   TunnelType = 4
)

func ParseTunnelType(t string) TunnelType {
	switch strings.ToLower(t) {
	case "tcp":
		return Tcp
	case "udp":
		return Udp
	case "socks":
		return Socks
	case "tun":
		return Tun
	case "tap":
		return Tap
	default:
		return Nil
	}
}

var config *StunningConfig = nil

func ReadConf() *StunningConfig {
	if config == nil {
		config = &StunningConfig{}
		config.SetConfigName("app")
		config.AddConfigPath(".")
		err := config.ReadInConfig()
		if err != nil {
			fmt.Println("Config file not found...")
			return nil
		}
	}
	config.tunConf = nil
	return config
}

func (c *StunningConfig) getApplicationConfs() *ApplicationConfig {
	conf := &ApplicationConfig{
		LogLevel:    c.GetInt("application.log_level"),
		LogLocation: c.GetString("application.log_location"),
		Crt:         c.GetString("application.crt"),
		Key:         c.GetString("application.key"),
	}
	return conf
}

func (c *StunningConfig) getTunnelConfs() map[string]*TunnelConfig {
	if c.tunConf == nil {
		c.tunConf = make(map[string]*TunnelConfig)
		for _, cname := range c.AllKeys() {
			tun := strings.Split(cname, ".")[0]
			if _, err := c.tunConf[tun]; !err {
				var conf *TunnelConfig = nil
				caddr := c.GetString(tun + ".connect")
				if caddr == "" {
					log.Fatalf("Tunnel %s has no Connect address config", tun)
					c.tunConf[tun] = conf
					break
				}

				laddr := c.GetString(tun + ".listen")
				if caddr == "" {
					log.Fatalf("Tunnel %s has no Listen address config", tun)
					c.tunConf[tun] = conf
					break
				}

				ttype := ParseTunnelType(c.GetString(tun + ".type"))
				if ttype == Nil {
					log.Fatalf("Tunnel %s has no Type config", tun)
					c.tunConf[tun] = conf
					break
				}
				conf = &TunnelConfig{
					ConnectAddr: caddr,
					ListenAddr:  laddr,
					TunnelType:  ttype,
					IsClient:    c.GetBool(tun + ".client"),
				}
				c.tunConf[tun] = conf
			}
		}
	}
}

func (c *StunningConfig) getTunnelConf(name string) *TunnelConfig {
	if val, err := c.tunConf[name]; !err {
		return val
	} else {
		return nil
	}
}
