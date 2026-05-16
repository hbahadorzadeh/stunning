package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hbahadorzadeh/stunning/core"
	"github.com/hbahadorzadeh/stunning/core/metrics"
)

const (
	version = "0.1.0"
)

type TunnelProcess struct {
	Name       string
	PID        int
	Config     core.TunnelConfig
	StartTime  time.Time
	Metrics    *metrics.Metrics
	MetricsURL string
}

var (
	configFile  = "tunnels.json"
	pidDir      = ""
	metricsPort = 9090
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	pidDir = filepath.Join(homeDir, ".stunning", "pids")
	os.MkdirAll(pidDir, 0755)
}

func readConfig(file string) (map[string]core.TunnelConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var confs map[string]core.TunnelConfig
	if err := json.Unmarshal(data, &confs); err != nil {
		return nil, err
	}
	return confs, nil
}

func getPIDFile(name string) string {
	return filepath.Join(pidDir, name+".pid")
}

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func startTunnel(name string, conf core.TunnelConfig, foreground bool) error {
	pidFile := getPIDFile(name)

	// Check if already running
	if data, err := os.ReadFile(pidFile); err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if isProcessAlive(pid) {
			return fmt.Errorf("tunnel %s already running (PID: %d)", name, pid)
		}
		os.Remove(pidFile)
	}

	// Get executable path
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Build command
	cmd := exec.Command(exe, "daemon-run", name, configFile)

	if foreground {
		// Foreground mode
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	// Background mode
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Save PID
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		cmd.Process.Kill()
		return err
	}

	fmt.Printf("✓ Tunnel '%s' started (PID: %d)\n", name, cmd.Process.Pid)
	return nil
}

func stopTunnel(name string) error {
	pidFile := getPIDFile(name)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("tunnel %s not running", name)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("tunnel %s (PID: %d) not found", name, pid)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	os.Remove(pidFile)
	fmt.Printf("✓ Tunnel '%s' stopped (PID: %d)\n", name, pid)
	return nil
}

func showStatus(configFile string) error {
	confs, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("no tunnels configured: %v", err)
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         Tunnel Status                                          ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-20s │ %-10s │ %-25s │ %-15s │\n", "Name", "Status", "Listen", "PID")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════╣")

	for name, conf := range confs {
		pidFile := getPIDFile(name)
		var status, pidStr string

		if data, err := os.ReadFile(pidFile); err == nil {
			pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
			if isProcessAlive(pid) {
				status = "✓ Running"
				pidStr = strconv.Itoa(pid)
			} else {
				status = "✗ Dead"
				pidStr = strconv.Itoa(pid)
				os.Remove(pidFile)
			}
		} else {
			status = "✗ Stopped"
			pidStr = "-"
		}

		fmt.Printf("║ %-20s │ %-10s │ %-25s │ %-15s │\n", name, status, conf.Listen, pidStr)
	}
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")

	fmt.Printf("\nMetrics available at: http://localhost:%d/metrics\n\n", metricsPort)
	return nil
}

func listTunnels(configFile string) error {
	confs, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("no configuration: %v", err)
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ Tunnels in %s%-60s║\n", configFile, "")
	fmt.Printf("╠════════════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ %-20s │ %-12s │ %-12s │ %-25s │ %-12s ║\n", "Name", "Mode", "Protocol", "Listen", "Connect")
	fmt.Printf("╠════════════════════════════════════════════════════════════════════════════════════════╣\n")

	for name, conf := range confs {
		fmt.Printf("║ %-20s │ %-12s │ %-12s │ %-25s │ %-12s ║\n",
			name, conf.ServiceMode, conf.ServerType, conf.Listen, conf.Connect)
	}
	fmt.Printf("╚════════════════════════════════════════════════════════════════════════════════════════╝\n\n")
	return nil
}

func daemonRun(tunnelName, cfgFile string) {
	confs, err := readConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	conf, exists := confs[tunnelName]
	if !exists {
		log.Fatalf("Tunnel %s not found in config", tunnelName)
	}

	// Start metrics HTTP server
	collector := metrics.GetGlobalCollector()
	httpServer, err := metrics.NewHTTPServer(fmt.Sprintf(":%d", metricsPort), collector)
	if err != nil {
		log.Printf("Warning: Could not start metrics server: %v", err)
	} else {
		defer httpServer.Close()
		log.Printf("Metrics available at http://localhost:%d/metrics", metricsPort)
	}

	// Create and run tunnel
	tun := core.TunnelFactory(tunnelName, conf)
	log.Printf("Starting tunnel: %s (PID: %d)", tunnelName, os.Getpid())

	tun.ListenAndServer()
}

func showHelp() {
	fmt.Printf(`Stunning - Network Tunneling CLI v%s

Manage network tunnels running as background processes.

Usage:
  stunning <command> [options]

Commands:
  start <name>               Start a tunnel in background
  fg <name>                  Run a tunnel in foreground (debugging)
  stop <name>                Stop a running tunnel
  status                     Show status of all tunnels
  list                       List all configured tunnels
  metrics                    Show Prometheus metrics
  version                    Show version
  help                       Show this help message

Options:
  -config <file>             Config file (default: tunnels.json)
  -metrics-port <port>       Metrics port (default: 9090)

Examples:
  # Start a tunnel from config
  stunning start my-tunnel

  # Run a tunnel in foreground (for debugging)
  stunning fg my-tunnel

  # Show all tunnel status
  stunning status

  # List configured tunnels
  stunning list

  # View Prometheus metrics
  curl http://localhost:9090/metrics

Default config: tunnels.json
Tunnel PIDs stored in: ~/.stunning/pids/

Configuration format (tunnels.json):
{
  "my-tunnel": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "tcp",
    "Listen": "127.0.0.1:8080",
    "Connect": "127.0.0.1:9090"
  }
}
`, version)
}

func main() {
	flag.StringVar(&configFile, "config", "tunnels.json", "Config file")
	flag.IntVar(&metricsPort, "metrics-port", 9090, "Metrics HTTP port")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		showHelp()
		return
	}

	command := args[0]

	// Handle daemon-run (internal, spawned by start/fg commands)
	if command == "daemon-run" {
		if len(args) < 3 {
			log.Fatalf("daemon-run requires tunnel name and config file")
		}
		daemonRun(args[1], args[2])
		return
	}

	switch command {
	case "start":
		if len(args) < 2 {
			fmt.Println("Error: tunnel name required")
			fmt.Println("Usage: stunning start <name>")
			return
		}
		name := args[1]
		confs, err := readConfig(configFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		conf, exists := confs[name]
		if !exists {
			fmt.Printf("Error: tunnel '%s' not found in config\n", name)
			return
		}
		if err := startTunnel(name, conf, false); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "fg":
		if len(args) < 2 {
			fmt.Println("Error: tunnel name required")
			fmt.Println("Usage: stunning fg <name>")
			return
		}
		name := args[1]
		confs, err := readConfig(configFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		conf, exists := confs[name]
		if !exists {
			fmt.Printf("Error: tunnel '%s' not found in config\n", name)
			return
		}
		if err := startTunnel(name, conf, true); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "stop":
		if len(args) < 2 {
			fmt.Println("Error: tunnel name required")
			fmt.Println("Usage: stunning stop <name>")
			return
		}
		if err := stopTunnel(args[1]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "status":
		if err := showStatus(configFile); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "list":
		if err := listTunnels(configFile); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "metrics":
		url := fmt.Sprintf("http://localhost:%d/metrics", metricsPort)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error: Could not reach metrics server at %s\n", url)
			fmt.Println("Make sure tunnels are running (use 'stunning status')")
			return
		}
		defer resp.Body.Close()
		fmt.Println(resp.Header.Get("Content-Type"))
		// Fetch and display metrics
		client := &http.Client{Timeout: 5 * time.Second}
		resp, _ = client.Get(url)
		defer resp.Body.Close()
		if _, err := os.Stdout.ReadFrom(resp.Body); err != nil {
			fmt.Printf("Error reading metrics: %v\n", err)
		}

	case "version":
		fmt.Printf("Stunning v%s\n", version)

	case "help", "-h", "--help":
		showHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'stunning help' for usage information")
	}
}
