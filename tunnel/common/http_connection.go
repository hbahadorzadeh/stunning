package common

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func GetCilentHttpConnection(proto, serverUrl string) (ClientHttpConnection, error) {
	var conn net.Conn
	var err error
	if proto == "http" {
		conn, err = net.Dial("tcp", serverUrl)
	} else {
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err = tls.Dial("tcp", serverUrl, conf)
	}
	if err != nil {
		return ClientHttpConnection{}, err
	}
	cd := ClientHttpDialer{
		Conn: conn,
	}
	tr := &http.Transport{Dial: cd.Dial}
	c := ClientHttpConnection{
		Connection: conn,
		Client: http.Client{
			Transport: tr,
		},
		ServerUrl: serverUrl,
		Ch:        make(chan []byte),
		RandomGen: rand.New(rand.NewSource(time.Now().UnixNano())),
		proto:     proto,
	}
	return c, nil
}

type ClientHttpDialer struct {
	proxy.Dialer
	Conn net.Conn
}

func (cl ClientHttpDialer) Dial(network, addr string) (c net.Conn, err error) {
	return cl.Conn, nil
}

type ClientHttpConnection struct {
	net.Conn
	Connection net.Conn
	Client     http.Client
	ServerUrl  string
	RandomGen  *rand.Rand
	Ch         chan []byte
	proto      string
}

func (c ClientHttpConnection) Read(b []byte) (n int, err error) {
	log.Printf("Waiting for Client bytes")
	buff := <-c.Ch
	n = len(b)
	if len(buff) > n {
		go func() { c.Ch <- buff[n:] }()
	} else {
		n = len(buff)
	}
	copy(b, buff[:n])
	log.Printf("reading Client bytes[%d](%v)", n, b)
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c ClientHttpConnection) Write(b []byte) (n int, err error) {
	var req *http.Request
	url := fmt.Sprintf("%s://%s/get?id=%d", "http", c.ServerUrl, c.RandomGen.Int63n(math.MaxInt64))
	if len(b) == 0 {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			log.Print(err)
			panic(err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-length", "0")
	} else {
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(b))
		if err != nil {
			log.Print(err)
			panic(err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-length", fmt.Sprint("%d", len(b)))
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}
	log.Printf("writing Client bytes %v -> %s", b, c.ServerUrl)
	go func() {
		buff := make([]byte, 4096)
		buff, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err != nil {
			log.Print(err)
			panic(err)
		}
		log.Printf("read Buffer %v", buff)
		c.Ch <- buff
		log.Printf("reading bytes(%v) %s ->", buff, c.ServerUrl)
	}()
	return len(b), err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c ClientHttpConnection) Close() error {
	return c.Connection.Close()
}

// LocalAddr returns the local network address.
func (c ClientHttpConnection) LocalAddr() net.Addr {
	return c.Connection.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c ClientHttpConnection) RemoteAddr() net.Addr {
	return c.Connection.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c ClientHttpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c ClientHttpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c ClientHttpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func GetServerHttpConnection() ServerHttpConnection {
	c := ServerHttpConnection{
		RCh:      make(chan []byte),
		WCh:      make(chan []byte),
		LastUsed: time.Now(),
		Closed:   false,
	}
	return c
}

type ServerHttpConnection struct {
	net.Conn
	RCh      chan []byte
	WCh      chan []byte
	LastUsed time.Time
	Closed   bool
}

func (c ServerHttpConnection) Read(b []byte) (n int, err error) {
	buff := <-c.RCh
	n = len(buff)
	b = b[:n]
	copy(b, buff)
	c.LastUsed = time.Now()
	log.Printf("Reading Server bytes %v(%v)", b, buff)
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c ServerHttpConnection) Write(b []byte) (n int, err error) {
	c.WCh <- b
	log.Printf("Writing Server bytes %v(%v)", b)
	return len(b), err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c ServerHttpConnection) Close() error {
	c.Closed = true
	return nil
}

// LocalAddr returns the local network address.
func (c ServerHttpConnection) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote network address.
func (c ServerHttpConnection) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c ServerHttpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c ServerHttpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c ServerHttpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
