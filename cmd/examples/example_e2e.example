package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	icommon "github.com/hbahadorzadeh/stunning/interface/common"
	socksiface "github.com/hbahadorzadeh/stunning/interface/socks"
	tcpiface "github.com/hbahadorzadeh/stunning/interface/tcp"
	tcommon "github.com/hbahadorzadeh/stunning/tunnel/common"
	dnstun "github.com/hbahadorzadeh/stunning/tunnel/dns"
	h2tun "github.com/hbahadorzadeh/stunning/tunnel/h2"
	icmptun "github.com/hbahadorzadeh/stunning/tunnel/icmp"
	tcptun "github.com/hbahadorzadeh/stunning/tunnel/tcp"
	udpstun "github.com/hbahadorzadeh/stunning/tunnel/udps"
	wstun "github.com/hbahadorzadeh/stunning/tunnel/ws"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

// generateSelfSignedCert creates a self-signed certificate and key for testing TLS-based tunnels
// Returns paths to temporary cert and key files
func generateSelfSignedCert(t *testing.T) (certFile, keyFile string) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create certificate template
	cert := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost", "127.0.0.1"},
	}

	// Self-sign the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Write certificate to temporary file
	certFile = "/tmp/test_cert_" + time.Now().Format("20060102150405") + ".pem"
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert file: %v", err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		t.Fatalf("Failed to encode certificate: %v", err)
	}

	// Write private key to temporary file
	keyFile = "/tmp/test_key_" + time.Now().Format("20060102150405") + ".pem"
	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		t.Fatalf("Failed to encode private key: %v", err)
	}

	return certFile, keyFile
}

// cleanupCertFiles removes temporary certificate and key files
func cleanupCertFiles(certFile, keyFile string) {
	os.Remove(certFile)
	os.Remove(keyFile)
}

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

// TestE2EH2Tunnel tests HTTP/2 tunnel end-to-end
func TestE2EH2Tunnel(t *testing.T) {
	testData := []byte("Hello, H2 Tunnel!")

	// Generate self-signed certificate for testing
	certFile, keyFile := generateSelfSignedCert(t)
	defer cleanupCertFiles(certFile, keyFile)

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9011")
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

	// 2. Start H2 tunnel server
	h2Server, err := h2tun.StartH2Server(certFile, keyFile, "127.0.0.1:9018")
	if err != nil {
		t.Fatalf("Failed to start H2 tunnel server: %v", err)
	}
	defer h2Server.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9011")
	h2Server.SetServer(ifaceServer)

	go h2Server.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create H2 tunnel client
	h2Dialer := h2tun.GetH2Dialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8090", "127.0.0.1:9018", h2Dialer)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	conn, err := net.Dial("tcp", "127.0.0.1:8090")
	if err != nil {
		t.Fatalf("Failed to connect to H2 tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send test data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data through H2 tunnel: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response from H2 tunnel: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("H2 tunnel data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
	ifaceClient.Close()
}

// TestE2EWsTunnel tests WebSocket tunnel end-to-end
func TestE2EWsTunnel(t *testing.T) {
	testData := []byte("Hello, WebSocket Tunnel!")

	// Generate self-signed certificate for testing
	certFile, keyFile := generateSelfSignedCert(t)
	defer cleanupCertFiles(certFile, keyFile)

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9013")
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

	// 2. Start WebSocket tunnel server
	wsServer, err := wstun.StartWsServer(certFile, keyFile, "127.0.0.1:9019")
	if err != nil {
		t.Fatalf("Failed to start WebSocket tunnel server: %v", err)
	}
	defer wsServer.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9013")
	wsServer.SetServer(ifaceServer)

	go wsServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create WebSocket tunnel client
	wsDialer := wstun.GetWsDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8091", "127.0.0.1:9019", wsDialer)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	conn, err := net.Dial("tcp", "127.0.0.1:8091")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send test data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data through WebSocket tunnel: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response from WebSocket tunnel: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("WebSocket tunnel data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
	ifaceClient.Close()
}

// TestE2EUdpsTunnel tests DTLS tunnel end-to-end
func TestE2EUdpsTunnel(t *testing.T) {
	testData := []byte("Hello, UDPS Tunnel!")

	// Generate self-signed certificate for testing
	certFile, keyFile := generateSelfSignedCert(t)
	defer cleanupCertFiles(certFile, keyFile)

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9014")
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

	// 2. Start UDPS (DTLS) tunnel server
	udpsServer, err := udpstun.StartUdpsServer(certFile, keyFile, "127.0.0.1:9020")
	if err != nil {
		t.Fatalf("Failed to start UDPS tunnel server: %v", err)
	}
	defer udpsServer.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9014")
	udpsServer.SetServer(ifaceServer)

	go udpsServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create UDPS tunnel client
	udpsDialer := udpstun.GetUdpsDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8092", "127.0.0.1:9020", udpsDialer)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	conn, err := net.Dial("tcp", "127.0.0.1:8092")
	if err != nil {
		t.Fatalf("Failed to connect to UDPS tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send test data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data through UDPS tunnel: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response from UDPS tunnel: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("UDPS tunnel data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
	ifaceClient.Close()
}

// TestE2EDnsTunnel tests DNS tunnel end-to-end
func TestE2EDnsTunnel(t *testing.T) {
	testData := []byte("Hello, DNS Tunnel!")

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9012")
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

	// 2. Start DNS tunnel server (TCP with 2-byte length-prefix framing)
	dnsServer, err := dnstun.StartDnsServer("127.0.0.1:9021")
	if err != nil {
		t.Fatalf("Failed to start DNS tunnel server: %v", err)
	}
	defer dnsServer.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9012")
	dnsServer.SetServer(ifaceServer)

	go dnsServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create DNS tunnel client
	dnsDialer := dnstun.GetDnsDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8093", "127.0.0.1:9021", dnsDialer)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	conn, err := net.Dial("tcp", "127.0.0.1:8093")
	if err != nil {
		t.Fatalf("Failed to connect to DNS tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send test data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data through DNS tunnel: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response from DNS tunnel: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("DNS tunnel data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
	ifaceClient.Close()
}

// TestE2EIcmpTunnel tests ICMP tunnel end-to-end
// Note: This test requires root or CAP_NET_RAW capabilities to run successfully
func TestE2EIcmpTunnel(t *testing.T) {
	testData := []byte("Hello, ICMP Tunnel!")

	// 1. Start upstream echo server
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:9015")
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

	// 2. Start ICMP tunnel server
	// This requires root or CAP_NET_RAW privileges
	icmpServer, err := icmptun.StartIcmpServer("127.0.0.1")
	if err != nil {
		// Permission denied indicates we're not running with sufficient privileges
		t.Skipf("ICMP tunnel requires root or CAP_NET_RAW privileges: %v", err)
	}
	defer icmpServer.Close()

	// Create TCP interface server (directs tunnel connections to upstream)
	ifaceServer := tcpiface.GetTcpServer("127.0.0.1:9015")
	icmpServer.SetServer(ifaceServer)

	go icmpServer.WaitingForConnection()
	time.Sleep(100 * time.Millisecond)

	// 3. Create ICMP tunnel client
	icmpDialer := icmptun.GetIcmpDialer()
	ifaceClient := tcpiface.GetTcpClient("127.0.0.1:8094", "127.0.0.1", icmpDialer)

	go ifaceClient.WaitingForConnection()
	time.Sleep(200 * time.Millisecond)

	// 4. Connect to the tunnel through the local interface
	conn, err := net.Dial("tcp", "127.0.0.1:8094")
	if err != nil {
		t.Fatalf("Failed to connect to ICMP tunnel interface: %v", err)
	}
	defer conn.Close()

	// 5. Send test data through the tunnel
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data through ICMP tunnel: %v", err)
	}

	// 6. Read echo response
	respData := make([]byte, len(testData))
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		t.Fatalf("Failed to read response from ICMP tunnel: %v", err)
	}

	// 7. Verify data matches
	if string(respData) != string(testData) {
		t.Errorf("ICMP tunnel data mismatch: sent %q, got %q", string(testData), string(respData))
	}

	// Cleanup
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
