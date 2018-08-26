package http

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type HttpServer struct {
	tcommon.TunnelServerCommon
	connMap   map[string]tcommon.ServerHttpConnection
	mux       sync.Mutex
	webserver *http.Server
	handler   func(http.ResponseWriter, *http.Request)
}

func StartHttpServer(address string) (HttpServer, error) {
	serv := HttpServer{
		connMap:   make(map[string]tcommon.ServerHttpConnection),
		webserver: &http.Server{Addr: address},
	}

	go func() {
		serv.mux.Lock()
		for range time.Tick(30 * time.Second) {
			toBeDeleted := make([]string, 0)
			for i, conn := range serv.connMap {
				if conn.LastUsed.Add(60*time.Second).Before(time.Now()) || conn.Closed {
					toBeDeleted = append(toBeDeleted, i)
				}
			}
			for _, i := range toBeDeleted {
				delete(serv.connMap, i)
			}
		}
		serv.mux.Unlock()
	}()

	return serv, nil
}

func (s HttpServer) SetServer(ss icommon.TunnelInterfaceServer) {
	s.Server = ss
	go s.WaitingForConnection()
	//time.Sleep(2 * time.Second)
}

func (s HttpServer) WaitingForConnection() {
	log.Println("Starting webserver")

	s.handler = func(w http.ResponseWriter, r *http.Request) {
		log.Print(r)
		var conn tcommon.ServerHttpConnection
		var err bool
		s.mux.Lock()
		if conn, err = s.connMap[r.RemoteAddr]; !err || conn.Closed {
			conn = tcommon.GetServerHttpConnection()
			s.connMap[r.RemoteAddr] = conn
			go s.HandleConnection(conn)
		}
		s.mux.Unlock()
		wbuff, ok := ioutil.ReadAll(r.Body)
		if ok == nil {
			conn.RCh <- wbuff
			//select {
			//case
			rbuff := <-conn.WCh
			w.Write(rbuff)
			//	break
			//default:
			//	break
			//}
		}
	}

	http.HandleFunc("/", s.handler)
	s.webserver.ListenAndServe()
}

func (s HttpServer) Close() error {
	s.webserver.Close()
	return nil
}

func (s HttpServer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	s.Server.HandleConnection(conn)
}
