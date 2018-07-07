package http

import (
	"bytes"
	"encoding/base64"
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

func getCilentHttpConnection(serverUrl string) (clientHttpConnection, error) {
	conn, err := net.Dial("tcp", serverUrl)
	if err != nil {
		return clientHttpConnection{}, err
	}
	cd := clientHttpDialer{
		conn: conn,
	}
	tr := &http.Transport{Dial: cd.Dial}
	c := clientHttpConnection{
		conn: conn,
		client: http.Client{
			Transport: tr,
		},
		serverUrl: serverUrl,
		ch:        make(chan []byte),
		rs:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return c, nil
}

type clientHttpDialer struct {
	proxy.Dialer
	conn net.Conn
}



func (cl clientHttpDialer) Dial(network, addr string) (c net.Conn, err error) {
	return cl.conn, nil
}

type clientHttpConnection struct {
	net.Conn
	conn      net.Conn
	client    http.Client
	serverUrl string
	rs        *rand.Rand
	ch        chan []byte
}

func (c clientHttpConnection) Read(b []byte) (n int, err error) {
	log.Printf("Waiting for client bytes")
	buff := <- c.ch
	n=len(b)
	if len(buff) > n{
		go func(){c.ch <- buff[n:]}()
	}
	copy(b, buff[:n])
	log.Printf("reading client bytes[%d](%v)", n, b)
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c clientHttpConnection) Write(b []byte) (n int, err error) {
	var req *http.Request
	url := fmt.Sprintf("http://%s/get?id=%d", c.serverUrl, c.rs.Int63n(math.MaxInt64))
	buff := make([]byte, 4096)
	if len(b) == 0 {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			log.Print(err)
			panic(err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-length", "0")
	} else {
		base64.StdEncoding.Encode(buff, b)
		buff = buff[:base64.StdEncoding.EncodedLen(len(b))]
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(buff))
		if err != nil {
			log.Print(err)
			panic(err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-length", fmt.Sprint("%d", len(buff)))
	}
	resp, err := c.client.Do(req)
	if err != nil {
		panic(err)
	}
	log.Printf("writing client bytes %v(%v) -> %s", b, buff, c.serverUrl)
	go func(){
		buff, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		rbuff := make([]byte, 4096)
		n,err = base64.StdEncoding.Decode(rbuff, buff)
		if err != nil {
			log.Print(err)
			panic(err)
		}
		rbuff = rbuff[:n]
		log.Printf("read buff %v", buff)
		log.Printf("decoded buff %v", rbuff)
		c.ch <- rbuff
		log.Printf("reading bytes(%v) %s ->", rbuff, c.serverUrl)
	}()
	return len(b), err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c clientHttpConnection) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
func (c clientHttpConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c clientHttpConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
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
func (c clientHttpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c clientHttpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c clientHttpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func getServerHttpConnection() serverHttpConnection {
	c := serverHttpConnection{
		rch:      make(chan []byte),
		wch:      make(chan []byte),
		lastUsed: time.Now(),
		isClosed: false,
	}
	return c
}

type serverHttpConnection struct {
	net.Conn
	rch      chan []byte
	wch      chan []byte
	lastUsed time.Time
	isClosed bool
}

func (c serverHttpConnection) Read(b []byte) (n int, err error) {
	buff := <-c.rch
	wb := make([]byte, 4096)
	n, err = base64.StdEncoding.Decode(wb, buff)
	if err != nil {
		log.Print(err)
		panic(err)
	}
	wb = wb[:n]
	copy(b, wb)
	c.lastUsed = time.Now()
	log.Printf("Reading Server bytes %v(%v)", b, buff)
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c serverHttpConnection) Write(b []byte) (n int, err error) {
	buff := make([]byte, 4096)
	base64.StdEncoding.Encode(buff, b)
	buff = buff[:base64.StdEncoding.EncodedLen(len(b))]
	c.wch <- buff
	log.Printf("Writing Server bytes %v(%v)", b, buff)
	return len(buff), err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c serverHttpConnection) Close() error {
	c.isClosed = true
	return nil
}

// LocalAddr returns the local network address.
func (c serverHttpConnection) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote network address.
func (c serverHttpConnection) RemoteAddr() net.Addr {
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
func (c serverHttpConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c serverHttpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c serverHttpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
