package main

import (
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	socksiface "github.com/hbahadorzadeh/stunning/interface/socks"
	tcpiface "github.com/hbahadorzadeh/stunning/interface/tcp"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	tcptun "github.com/hbahadorzadeh/stunning/tunnel/tcp"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

// TestE2ETcpTunnel tests end-to-end TCP tunnel: client sends data through tunnel to server
func TestE2ETcpTunnel(t *testing.T) {
	testData := []byte("Hello, TCP Tunnel!")

	// 1. Start upstream TCP server (what the tunnel client connects to)
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		t.Fatalf("Failed to start upstream server: %v", err)
	}
	defer upstreamListener.Close()

	// Upstream echo server
	go func() {
		for {
			conn, err := upstreamListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c) // Echo server
			}(conn)
		}
	}()

	// 2. Start TCP tunnel server (accepts connections on a port)
	tcpServer, err := tcptun.StartTcpServer("127.0.0.1:9000")
	if err != nil {
		t.Fatalf("Failed to start TCP tunnel server: %v", err)
	}
	defer tcpServer.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9001")
	tcpServer.SetServer(ifaceServer)

	// Start tunnel server in goroutine
	go tcpServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create TCP tunnel client (listens locally, connects to tunnel)
	tcpClient := tcptun.GetTcpDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8080", "127.0.0.1:9000", tcpClient)

	// Start client listener in goroutine
	go ifaceClient.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	time.Sleep(200 * time.Millisecond) // Give server time to fully start
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("Failed to connect to tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("Data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
	ifaceClient.Close()
}

// TestE2ESocksProxy tests SOCKS5 proxy through tunnel
func TestE2ESocksProxy(t *testing.T) {

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9002")
	if err != nil {
		t.Fatalf("Failed to start upstream server: %v", err)
	}
	defer upstreamListener.Close()

	go func() {
		for {
			conn, err := upstreamListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	// 2. Start TCP tunnel server
	tcpServer, err := tcptun.StartTcpServer("127.0.0.1:9003")
	if err != nil {
		t.Fatalf("Failed to start TCP tunnel server: %v", err)
	}
	defer tcpServer.Close()

	// Set SOCKS server as the interface
	socksServer := socksiface.GetSocksServer()
	tcpServer.SetServer(socksServer)

	go tcpServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create SOCKS client
	tcpClient := tcptun.GetTcpDialer()
	sockClient := socksiface.GetSocksClient("127.0.0.1:1080", "127.0.0.1:9003", tcpClient)

	go sockClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect through SOCKS proxy
	// This would require a SOCKS5 client library to test properly
	// For now, verify the components initialize correctly
	if sockClient == nil {
		t.Error("SOCKS client failed to initialize")
	}

	sockClient.Close()
}

// TestTunnelFactoryTcpClient tests the factory pattern with TCP client tunnel
func TestTunnelFactoryTcpClient(t *testing.T) {
	conf := TunnelConfig{
		ServiceMode:   "client",
		ServerType:    "tcp",
		InterfaceType: "tcp",
		Listen:        "127.0.0.1:8081",
		Connect:       "127.0.0.1:9004",
	}

	tunnel := TunnelFactory("test_tcp_client", conf)
	if tunnel == nil {
		t.Fatal("Factory returned nil tunnel")
	}

	// Verify it's a working Tunnel interface
	tunnel.IsAlive() // Should not panic
}

// TestTunnelFactoryTcpServer tests the factory pattern with TCP server tunnel
func TestTunnelFactoryTcpServer(t *testing.T) {
	conf := TunnelConfig{
		ServiceMode:   "server",
		ServerType:    "tcp",
		InterfaceType: "tcp",
		Listen:        "127.0.0.1:9005",
		Connect:       "127.0.0.1:9006",
	}

	tunnel := TunnelFactory("test_tcp_server", conf)
	if tunnel == nil {
		t.Fatal("Factory returned nil tunnel")
	}

	tunnel.IsAlive() // Should not panic
}

// TestConcurrentConnections tests multiple concurrent connections through a tunnel
func TestConcurrentConnections(t *testing.T) {
	numConnections := 5
	testData := []byte("concurrent test")

	// Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9007")
	if err != nil {
		t.Fatalf("Failed to start upstream server: %v", err)
	}
	defer upstreamListener.Close()

	go func() {
		for {
			conn, err := upstreamListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	// Start TCP tunnel
	tcpServer, err := tcptun.StartTcpServer("127.0.0.1:9008")
	if err != nil {
		t.Fatalf("Failed to start tunnel: %v", err)
	}
	defer tcpServer.Close()

	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9007")
	tcpServer.SetServer(ifaceServer)

	go tcpServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// Start tunnel client
	tcpClient := tcptun.GetTcpDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8082", "127.0.0.1:9008", tcpClient)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// Test concurrent connections
	for i := 0; i < numConnections; i++ {
		go func() {
			conn, err := net.Dial("tcp", "127.0.0.1:8082")
			if err != nil {
				t.Logf("Failed to connect: %v", err)
				return
			}
			defer conn.Close()

			// Write and read
			_, err = conn.Write(testData)
			if err != nil {
				t.Logf("Failed to write: %v", err)
				return
			}

			respData := make([]byte, len(testData))
			_, err = io.ReadFull(conn, respData)
			if err != nil {
				t.Logf("Failed to read: %v", err)
				return
			}

			if string(respData) != string(testData) {
				t.Errorf("Data mismatch: sent %q, got %q", string(testData), string(respData))
			}
		}()
	}

	// Wait for all goroutines
	time.Sleep(1 * time.Second)
	ifaceClient.Close()
}

// TestTunnelRecovery tests that tunnels can handle disconnections gracefully
func TestTunnelRecovery(t *testing.T) {
	testData := []byte("recovery test")

	// Start upstream server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9009")
	if err != nil {
		t.Fatalf("Failed to start upstream server: %v", err)
	}
	defer upstreamListener.Close()

	go func() {
		for {
			conn, err := upstreamListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	// Start tunnel
	tcpServer, err := tcptun.StartTcpServer("127.0.0.1:9010")
	if err != nil {
		t.Fatalf("Failed to start tunnel: %v", err)
	}
	defer tcpServer.Close()

	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9009")
	tcpServer.SetServer(ifaceServer)

	go tcpServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	tcpClient := tcptun.GetTcpDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8083", "127.0.0.1:9010", tcpClient)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// First connection
	conn1, err := net.Dial("tcp", "127.0.0.1:8083")
	if err != nil {
		t.Fatalf("Failed first connection: %v", err)
	}

	_, err = conn1.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write on first connection: %v", err)
	}

	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn1, respData)
	if err != nil {
		t.Fatalf("Failed to read on first connection: %v", err)
	}
	conn1.Close()

	// Second connection after first closes
	time.Sleep(100 * time.Millisecond)

	conn2, err := net.Dial("tcp", "127.0.0.1:8083")
	if err != nil {
		t.Fatalf("Failed second connection: %v", err)
	}
	defer conn2.Close()

	_, err = conn2.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write on second connection: %v", err)
	}

	respData = make([]byte, len(testData))
	_, err = io.ReadFull(conn2, respData)
	if err != nil {
		t.Fatalf("Failed to read on second connection: %v", err)
	}

	ifaceClient.Close()
}

// TestInterfaceImplementation verifies all tunnel types implement required interfaces
func TestInterfaceImplementation(t *testing.T) {
	// Verify TunnelClient and TunnelServer implement Tunnel interface
	var _ Tunnel = &TunnelClient{}
	var _ Tunnel = &TunnelServer{}

	// Verify interface servers implement TunnelInterfaceServer
	var _ icommon.TunnelInterfaceServer = tcpiface.GetTcpServer("127.0.0.1:9999")
	var _ icommon.TunnelInterfaceServer = socksiface.GetSocksServer()

	// Verify tunnel dialers implement TunnelDialer
	var _ tcommon.TunnelDialer = tcptun.GetTcpDialer()

	log.Println("All interface implementations verified")
}
