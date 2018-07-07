package tun

import (
	"github.com/songgao/water"
	"gitlab.com/h.bahadorzadeh/stunning/common"
	icommon "gitlab.com/h.bahadorzadeh/stunning/interface/common"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"log"
	"net"
	"os"
	"os/exec"
)

type TunInterface struct {
	icommon.TunnelInterfaceServer
	icommon.TunnelInterfaceClient
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

func GetTunIface(config TunConfig) TunInterface {
	ifce, err := water.New(water.Config{
		DeviceType: config.DevType,
	})

	runIP("link", "set", "dev", ifce.Name(), "mtu", config.MTU)
	runIP("addr", "add", config.Address, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if err != nil {
		log.Fatal(err)
	}

	iface := TunInterface{
		iface: ifce,
		conf:  config,
	}
	return iface
}

func GetTunIfaceClient(config TunConfig, addr string, d tcommon.TunnelDialer) TunInterface {
	ifce, err := water.New(water.Config{
		DeviceType: config.DevType,
	})

	runIP("link", "set", "dev", ifce.Name(), "mtu", config.MTU)
	runIP("addr", "add", config.Address, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if err != nil {
		log.Fatal(err)
	}

	iface := TunInterface{
		iface: ifce,
		conf:  config,
	}
	conn, err := d.Dial(d.Protocol().String(), addr)
	go iface.HandleConnection(conn)
	return iface
}

func (t TunInterface) HandleConnection(conn net.Conn) error {
	log.Printf("Tun iface %s handling new connection \n", t.iface.Name())
	go t.reader(conn)
	t.writer(conn)
	return nil
}

func (t *TunInterface) reader(conn net.Conn) {
	var frame common.Frame

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

func (t *TunInterface) writer(conn net.Conn) {
	var frame common.Frame
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
