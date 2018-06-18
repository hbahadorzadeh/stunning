package lib

import (
	"crypto/tls"
	"log"
	"net"
)

func StartTlsServer() {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":4443", config)

	if err != nil {
		log.Println(err)
		return
	}
	//defer ln.Close()

	sserver := get_socks_server()
	log.Println("socks server created")
	go waiting_for_connection(ln, sserver)
	log.Println("Done tls server")
}

func waiting_for_connection(ln net.Listener, sserver *socks_server){
	for {
		conn, err := ln.Accept()
		log.Println("new connection")
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, sserver)
	}
}

func handleConnection(conn net.Conn, sserver *socks_server) {
	defer conn.Close()
	sserver.handle_connection(conn)
}
