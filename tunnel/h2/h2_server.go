package h2

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	"log"
	"net/http"
	"sync"
	"time"
)

type H2Server struct {
	tcommon.TunnelServerCommon
	connMap        map[string]tcommon.ServerHttpConnection
	mux            sync.Mutex
	webserver      *http.Server
	mux_handler    *http.ServeMux
	crt, key       string
	cleanupDone    chan struct{}
	cleanupTicker  *time.Ticker
}

func StartH2Server(crt, key, address string) (*H2Server, error) {
	serv := &H2Server{
		connMap:       make(map[string]tcommon.ServerHttpConnection),
		webserver:     &http.Server{Addr: address},
		mux_handler:   http.NewServeMux(),
		crt:           crt,
		key:           key,
		cleanupDone:   make(chan struct{}),
		cleanupTicker: time.NewTicker(30 * time.Second),
	}

	go func() {
		defer serv.cleanupTicker.Stop()
		for {
			select {
			case <-serv.cleanupTicker.C:
				serv.mux.Lock()
				toBeDeleted := make([]string, 0)
				for i, conn := range serv.connMap {
					if conn.LastUsed.Add(60*time.Second).Before(time.Now()) || conn.Closed {
						toBeDeleted = append(toBeDeleted, i)
					}
				}
				for _, i := range toBeDeleted {
					delete(serv.connMap, i)
				}
				serv.mux.Unlock()
			case <-serv.cleanupDone:
				return
			}
		}
	}()

	return serv, nil
}

func (s *H2Server) SetServer(server icommon.TunnelInterfaceServer) {
	s.Server = server
	go s.WaitingForConnection()
}

func (s *H2Server) WaitingForConnection() {
	s.mux_handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		addr := r.RemoteAddr
		var buf [512]byte
		n, _ := r.Body.Read(buf[:])
		data := make([]byte, n)
		copy(data, buf[:n])

		s.mux.Lock()
		conn, exists := s.connMap[addr]
		if !exists || conn.Closed {
			conn = tcommon.GetServerHttpConnection()
			s.connMap[addr] = conn
			connPtr := &s.connMap[addr]
			s.mux.Unlock()

			go s.HandleConnection(connPtr)
			connPtr.RCh <- data
		} else {
			s.mux.Unlock()
			conn.RCh <- data
		}

		result := <-conn.WCh
		w.Write(result)
	})

	s.webserver.Handler = s.mux_handler
	if err := s.webserver.ListenAndServeTLS(s.crt, s.key); err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
}

func (s *H2Server) Close() error {
	// Close the webserver first to stop accepting new connections
	err := s.webserver.Close()

	// Signal cleanup goroutine to exit
	close(s.cleanupDone)

	// Close all connection channels to unblock handlers
	s.mux.Lock()
	for _, conn := range s.connMap {
		close(conn.RCh)
		close(conn.WCh)
	}
	s.mux.Unlock()

	return err
}
