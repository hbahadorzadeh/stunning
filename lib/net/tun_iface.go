package net

import (
	"github.com/songgao/water"
	"log"
	"net"
	"os"
	"os/exec"
)

type tun_interface struct {
	Vpnserver
	conf  TunConfig
	iface *water.Interface
	//nat map[net.IP]
}

type TunConfig struct {
	DevType water.DeviceType
	Address string
	Name    string
	MTU     string
}

func GetTunIface(config TunConfig) *tun_interface {
	ifce, err := water.New(water.Config{
		DeviceType: config.DevType,
	})

	runIP("link", "set", "dev", ifce.Name(), "mtu", config.MTU)
	runIP("addr", "add", config.Address, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if err != nil {
		log.Fatal(err)
	}

	iface := &tun_interface{
		iface: ifce,
		conf:  config,
	}
	return iface
}

func (t *tun_interface) HandleConnection(conn net.Conn) error {
	log.Printf("Tun iface %s handling new connection \n", t.iface.Name())
	go t.reader(conn)
	t.writer(conn)
	return nil
}

func (t *tun_interface) reader(conn net.Conn) {
	var frame Frame

	for {
		frame.Resize(1500)
		n, err := t.iface.Read([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}
		frame = frame[:n]
		if len(frame) == 0 {
			continue
		}
		wn, werr := conn.Write(frame)
		if werr != nil || wn != len(frame) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%d bytes wrote to socket", wn)
	}
}

func (t *tun_interface) writer(conn net.Conn) {
	var frame Frame
	for {
		frame.Resize(1500)
		n, err := conn.Read([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}
		frame = frame[:n]
		if len(frame) == 0 {
			continue
		}
		wn, werr := t.iface.Write(frame)
		if err != nil || wn != len(frame) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%d bytes wrote to iface", wn)
	}
}

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Fatalln("Error running /sbin/ip:", err)
	}
}
