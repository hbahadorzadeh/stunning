package h2

import (
	"log"
	"net/http"
	"sync"
	"time"

	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
)

type H2Server struct {
	tcommon.TunnelServerCommon
	connMap       map[string]*tcommon.ServerHttpConnection
	mux           sync.Mutex
	webserver     *http.Server
	mux_handler   *http.ServeMux
	crt, key      string
	stopCleanup   chan struct{}
	cleanupTicker *time.Ticker
}

func StartH2Server(crt, key, address string) (*H2Server, error) {
	serv := &H2Server{
		connMap:       make(map[string]*tcommon.ServerHttpConnection),
		webserver:     &http.Server{Addr: address},
		mux_handler:   http.NewServeMux(),
		crt:           crt,
		key:           key,
		stopCleanup:   make(chan struct{}),
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
					if conn != nil && (conn.LastUsed.Add(60*time.Second).Before(time.Now()) || conn.Closed) {
						toBeDeleted = append(toBeDeleted, i)
					}
				}
				for _, i := range toBeDeleted {
					delete(serv.connMap, i)
				}
				serv.mux.Unlock()
			case <-serv.stopCleanup:
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
		s.mux.Lock()
		conn, exists := s.connMap[addr]
		s.mux.Unlock()

		if !exists || conn == nil || conn.Closed {
			var buf [512]byte
			n, _ := r.Body.Read(buf[:])
			data := make([]byte, n)
			copy(data, buf[:n])

			connVal := tcommon.GetServerHttpConnection()
			s.mux.Lock()
			s.connMap[addr] = &connVal
			s.mux.Unlock()

			connRef := s.connMap[addr]
			go s.HandleConnection(connRef)
			connRef.RCh <- data
		} else {
			var buf [512]byte
			n, _ := r.Body.Read(buf[:])
			data := make([]byte, n)
			copy(data, buf[:n])
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
	// Stop cleanup goroutine
	close(s.stopCleanup)

	// Close webserver to stop accepting new connections
	err := s.webserver.Close()

	// Close all connection channels to unblock handlers
	s.mux.Lock()
	for _, conn := range s.connMap {
		close(conn.RCh)
		close(conn.WCh)
	}
	s.mux.Unlock()

	return err
}
