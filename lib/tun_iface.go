package lib

import (
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"log"
	"net"
	"os/exec"
	"os"
)

type TunInterface struct {
	Vpnserver
	conf TunConfig
	iface *water.Interface
	//nat map[net.IP]
}

type TunConfig struct {
	DevType water.DeviceType
	Address string
	Name    string
	MTU     string
}

func GetTunIface(config TunConfig) *TunInterface {
	ifce, err := water.New(water.Config{
		DeviceType: config.DevType,
	})

	runIP("link", "set", "dev", ifce.Name(), "mtu", config.MTU)
	runIP("addr", "add", config.Address, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if err != nil {
		log.Fatal(err)
	}

	iface := &TunInterface{
		iface: ifce,
		conf: config,
	}
	return iface
}

func (t *TunInterface) HandleConnection(conn net.Conn) error {
	log.Printf("Tun iface %s handling new connection \n" , t.iface.Name())
	go t.Reader(conn)
	t.Writer(conn)
	return nil
}

func (t *TunInterface) Reader(conn net.Conn) {
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
		log.Printf("%d bytes read from iface", n)
		log.Printf("Dst: %s\n", frame.Destination())
		log.Printf("Src: %s\n", frame.Source())
		log.Printf("Ethertype: % x\n", frame.Ethertype())
		log.Printf("Payload: % x\n", frame.Payload())
		wn, werr := conn.Write(frame)
		if werr != nil || wn != len(frame) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%d bytes wrote to socket", wn)
	}
}

func (t *TunInterface) Writer(conn net.Conn) {
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
		waterutil.SetIPv4Source(frame, net.ParseIP(t.conf.Address))
		log.Printf("%d bytes read from socket", n)
		log.Printf("Dst: %s\n", frame.Destination())
		log.Printf("Src: %s\n", frame.Source())
		log.Printf("Ethertype: % x\n", frame.Ethertype())
		log.Printf("Payload: % x\n", frame.Payload())
		wn, werr := t.iface.Write(frame)
		if err != nil || wn != len(frame) {
			log.Panicln(werr)
			log.Printf("wn : %d, n: %d \n", wn, n)
		}
		log.Printf("%d bytes wrote to iface", wn)
	}
}

func runIP(args... string) {
	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Fatalln("Error running /sbin/ip:", err)
	}
}