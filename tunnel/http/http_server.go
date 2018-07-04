package http

import (
	"net/http"
	tcommon "gitlab.com/h.bahadorzadeh/stunning/tunnel/common"
	"sync"
	"net"
	"time"
	"io/ioutil"
	"log"
)


type HttpServer struct {
	tcommon.TunnelServer
	connMap map[string]serverHttpConnection
	mux     sync.Mutex
	server *http.Server
}

func StartHttpServer(address string) (HttpServer, error) {
	serv := HttpServer{
		connMap: make(map[string]serverHttpConnection),
		server : &http.Server{Addr:address},
	}

	http.HandleFunc("/", serv.handler)

	go func() {
		serv.mux.Lock()
		for range time.Tick(30 * time.Second) {
			toBeDeleted := make([]string, 0)
			for i, conn := range serv.connMap {
				if conn.lastUsed.Add(60*time.Second).Before(time.Now()) || conn.isClosed {
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

func(s HttpServer)WaitingForConnection(){
	s.server.ListenAndServe()
}


func(s HttpServer)Close(){
	s.server.Close()
}

func(s HttpServer) handler(w http.ResponseWriter, r *http.Request) {
	log.Print(r)
	var conn serverHttpConnection
	var err bool
	s.mux.Lock()
	if conn, err = s.connMap[r.RemoteAddr]; !err  || conn.isClosed{
		conn = getServerHttpConnection()
		go s.HandleConnection(conn)
	}
	s.mux.Unlock()
	wbuff, ok := ioutil.ReadAll(r.Body)
	if ok == nil {
		conn.rch <- wbuff
		if len(conn.wch) > 0 {
			rbuff := <- conn.wch
			w.Write(rbuff)
		}
	}
}

func (s HttpServer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	s.Server.HandleConnection(conn)
}
