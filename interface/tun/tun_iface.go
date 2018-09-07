package tun

import (
	"github.com/hbahadorzadeh/stunning/common"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"github.com/songgao/water"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

type TunInterface struct {
	icommon.TunnelInterfaceServer
	icommon.TunnelInterfaceClient
	conf    TunConfig
	iface   *water.Interface
	stopped bool
}

type TunInterfaceClient struct {
	TunInterface
	address string
	dialer  tcommon.TunnelDialer
	closed  bool
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
		iface:   ifce,
		conf:    config,
		stopped: false,
	}
	return iface
}

func GetTunIfaceClient(config TunConfig, addr string, d tcommon.TunnelDialer) TunInterfaceClient {
	ifce, err := water.New(water.Config{
		DeviceType: config.DevType,
	})

	runIP("link", "set", "dev", ifce.Name(), "mtu", config.MTU)
	runIP("addr", "add", config.Address, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if err != nil {
		log.Fatal(err)
	}

	iface := TunInterfaceClient{
		address: addr,
		dialer:  d,
	}
	iface.iface = ifce
	iface.conf = config
	iface.closed = false
	return iface
}

func (t TunInterface) HandleConnection(conn net.Conn) error {
	log.Printf("Tun iface %s handling new connection \n", t.iface.Name())
	go t.reader(conn)
	t.writer(conn)
	return nil
}

func (t TunInterface) reader(conn net.Conn) {
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

func (t TunInterface) writer(conn net.Conn) {
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

func (t *TunInterfaceClient) WaitingForConnection() {
	conn, err := t.dialer.Dial(t.dialer.Protocol().String(), t.address)
	if err == nil {
		t.HandleConnection(conn)
	}
	t.closed = true
}

func (t *TunInterfaceClient) Close() {
	t.iface.Close()
	t.closed = true
}

func (t TunInterfaceClient) Closed() bool {
	return t.closed
}

func (t TunInterface) WaitingForConnection() {
	for !t.stopped {
		time.Sleep(time.Second)
	}
}

func (t TunInterface) Close() error {
	t.stopped = true
	return nil
}
