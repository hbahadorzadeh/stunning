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
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/hbahadorzadeh/stunning/core"
	"github.com/hbahadorzadeh/stunning/core/metrics"
)

const version = "1.1.1"

var (
	configFile  = "tunnels.json"
	pidDir      = ""
	metricsPort = 9090
)

// --- styling -----------------------------------------------------------------

var (
	colorPrimary = lipgloss.Color("63")  // indigo - brand
	colorAccent  = lipgloss.Color("212") // pink - flags/highlights
	colorOK      = lipgloss.Color("42")  // green
	colorErr     = lipgloss.Color("203") // red
	colorDim     = lipgloss.Color("241") // gray

	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	cmdStyle     = lipgloss.NewStyle().Foreground(colorPrimary)
	flagStyle    = lipgloss.NewStyle().Foreground(colorAccent)
	dimStyle     = lipgloss.NewStyle().Foreground(colorDim)
	okStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorOK)
	errStyle     = lipgloss.NewStyle().Bold(true).Foreground(colorErr)
)

func printOK(format string, a ...any) {
	fmt.Println(okStyle.Render("✓") + " " + fmt.Sprintf(format, a...))
}

func printErr(format string, a ...any) {
	fmt.Fprintln(os.Stderr, errStyle.Render("✗ Error:")+" "+fmt.Sprintf(format, a...))
}

func fail(format string, a ...any) {
	printErr(format, a...)
	os.Exit(1)
}

// kvTable renders a borderless two-column table. Styling individual cells
// (e.g. coloring a flag name) does not break column alignment, unlike
// fmt.Printf("%-Ns", ...) which counts ANSI escape bytes as width.
func kvTable(rows [][2]string, keyStyle lipgloss.Style) string {
	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderRow(false).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col == 0 {
				return keyStyle.PaddingRight(2)
			}
			return dimStyle
		})
	for _, r := range rows {
		t.Row(r[0], r[1])
	}
	return t.Render()
}

// --- init ----------------------------------------------------------------

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

// --- ad-hoc tunnel config (no tunnels.json required) ----------------------

// tunnelFlags holds tunnel-config overrides passed directly on the command
// line. Combined with resolveConfig, this lets `fg`/`start` run without any
// entry in tunnels.json — or override individual fields from one.
type tunnelFlags struct {
	mode, protocol, iface, listen, connect, plugins, auth, knock, key, cert, device, mtu string
}

// resolveConfig builds the TunnelConfig for `name`, starting from
// tunnels.json (if present and it has this entry) and then applying any
// non-empty command-line overrides on top.
func resolveConfig(name string, f tunnelFlags) (core.TunnelConfig, error) {
	var conf core.TunnelConfig
	if confs, err := readConfig(configFile); err == nil {
		conf = confs[name]
	}

	if f.mode != "" {
		conf.ServiceMode = f.mode
	}
	if f.protocol != "" {
		conf.ServerType = f.protocol
	}
	if f.iface != "" {
		conf.InterfaceType = f.iface
	}
	if f.listen != "" {
		conf.Listen = f.listen
	}
	if f.connect != "" {
		conf.Connect = f.connect
	}
	if f.plugins != "" {
		conf.Plugins = f.plugins
	}
	if f.auth != "" {
		conf.Auth = f.auth
	}
	if f.knock != "" {
		conf.Knock = f.knock
	}
	if f.key != "" {
		conf.Key = f.key
	}
	if f.cert != "" {
		conf.Cert = f.cert
	}
	if f.device != "" {
		conf.DeviceName = f.device
	}
	if f.mtu != "" {
		conf.Mtu = f.mtu
	}

	if conf.ServiceMode == "" || conf.ServerType == "" || conf.InterfaceType == "" || conf.Listen == "" {
		return conf, fmt.Errorf("tunnel '%s' not found in %s, and not enough flags given\n  need at least: -mode, -protocol, -iface, -listen", name, configFile)
	}
	return conf, nil
}

// --- tunnel process management --------------------------------------------

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

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Persist the fully-resolved config (tunnels.json entry + CLI overrides)
	// so daemon-run sees exactly what was requested, even with no config file.
	runConfig := filepath.Join(pidDir, name+".run.json")
	data, err := json.Marshal(map[string]core.TunnelConfig{name: conf})
	if err != nil {
		return err
	}
	if err := os.WriteFile(runConfig, data, 0600); err != nil {
		return err
	}

	cmd := exec.Command(exe, "daemon-run", name, runConfig)

	if foreground {
		// Foreground mode
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	// Background mode: detach the daemon (platform-specific, see procattr_*.go).
	cmd.SysProcAttr = detachSysProcAttr()

	if err := cmd.Start(); err != nil {
		return err
	}

	// Save PID
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		cmd.Process.Kill()
		return err
	}

	printOK("Tunnel '%s' started (PID: %d)", name, cmd.Process.Pid)
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
	printOK("Tunnel '%s' stopped (PID: %d)", name, pid)
	return nil
}

func showStatus(configFile string) error {
	confs, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("no tunnels configured in %s: %w", configFile, err)
	}

	names := make([]string, 0, len(confs))
	for name := range confs {
		names = append(names, name)
	}
	sort.Strings(names)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(dimStyle).
		Headers("NAME", "STATUS", "LISTEN", "PID").
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headingStyle.Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	for _, name := range names {
		conf := confs[name]
		pidFile := getPIDFile(name)
		var status, pidStr string

		if data, err := os.ReadFile(pidFile); err == nil {
			pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
			if isProcessAlive(pid) {
				status = okStyle.Render("✓ running")
				pidStr = strconv.Itoa(pid)
			} else {
				status = errStyle.Render("✗ dead")
				pidStr = strconv.Itoa(pid)
				os.Remove(pidFile)
			}
		} else {
			status = dimStyle.Render("- stopped")
			pidStr = "-"
		}

		t.Row(name, status, conf.Listen, pidStr)
	}

	fmt.Println(titleStyle.Render("Tunnel Status"))
	fmt.Println(t.Render())
	fmt.Println(dimStyle.Render(fmt.Sprintf("Metrics: http://localhost:%d/metrics", metricsPort)))
	return nil
}

func listTunnels(configFile string) error {
	confs, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("no configuration in %s: %w", configFile, err)
	}

	names := make([]string, 0, len(confs))
	for name := range confs {
		names = append(names, name)
	}
	sort.Strings(names)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(dimStyle).
		Headers("NAME", "MODE", "PROTOCOL", "LISTEN", "CONNECT").
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headingStyle.Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	for _, name := range names {
		conf := confs[name]
		t.Row(name, conf.ServiceMode, conf.ServerType, conf.Listen, conf.Connect)
	}

	fmt.Println(titleStyle.Render(fmt.Sprintf("Tunnels in %s", configFile)))
	fmt.Println(t.Render())
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

// --- help -------------------------------------------------------------------

func showHelp() {
	h := headingStyle.Render
	c := cmdStyle.Render
	d := dimStyle.Render

	fmt.Println(titleStyle.Render(fmt.Sprintf("Stunning v%s", version)) + d("  — anti-DPI plugin-chain tunnel"))
	fmt.Println(d("Manage network tunnels running as background processes."))
	fmt.Println()

	fmt.Println(h("Usage"))
	fmt.Printf("  stunning [flags] %s [args]\n\n", d("<command>"))

	fmt.Println(h("Commands"))
	fmt.Println(kvTable([][2]string{
		{"start <name>", "Start a tunnel in the background"},
		{"fg <name>", "Run a tunnel in the foreground (debug)"},
		{"stop <name>", "Stop a running tunnel"},
		{"status", "Show status of all tunnels"},
		{"list", "List configured tunnels"},
		{"metrics", "Print Prometheus metrics"},
		{"version", "Show version"},
		{"help", "Show this help"},
	}, cmdStyle))
	fmt.Println()

	fmt.Println(h("Global flags"))
	fmt.Println(kvTable([][2]string{
		{"-config <file>", `tunnels.json file (default "tunnels.json")`},
		{"-metrics-port <port>", "metrics HTTP port (default 9090)"},
	}, flagStyle))
	fmt.Println()

	fmt.Println(h("Run without a config file") + d(" (fg / start, or to override a tunnels.json entry)"))
	fmt.Println(kvTable([][2]string{
		{"-mode <server|client>", "tunnel mode"},
		{"-protocol <tcp|udp|tls|h2|ws|udps|dns|icmp|http|https>", "wire protocol"},
		{"-iface <tcp|udp|socks|tun|serial>", "local interface"},
		{"-listen <addr>", "address to listen on, e.g. :8443"},
		{"-connect <addr>", "address to forward to / dial"},
		{"-plugins <chain>", "anti-DPI plugin chain, e.g. aead?key=...,tls-mimic"},
		{"-auth <spec>", "auth gate: psk|jwt|mtls|oauth|ldap"},
		{"-knock <spec>", "port-knock (spa) gate"},
		{"-cert <file> / -key <file>", "TLS certificate / key"},
		{"-device <name> / -mtu <n>", "TUN device name / MTU"},
	}, flagStyle))
	fmt.Println()

	fmt.Println(h("Examples"))
	fmt.Println(d("  # from tunnels.json"))
	fmt.Printf("  stunning fg %s\n\n", c("my-tunnel"))
	fmt.Println(d("  # no config file needed"))
	fmt.Println("  stunning -mode server -protocol tcp -iface tcp \\")
	fmt.Println("    -listen :8443 -connect 127.0.0.1:9000 \\")
	fmt.Printf("    -plugins \"aead?key=...,tls-mimic\" fg %s\n\n", c("edge"))
	fmt.Println("  stunning status")
	fmt.Println("  stunning list")
	fmt.Println("  curl http://localhost:9090/metrics")
	fmt.Println()

	fmt.Println(h("Config file") + d(" (tunnels.json)"))
	fmt.Println(dimStyle.Render(`{
  "my-tunnel": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "tcp",
    "Listen": "127.0.0.1:8080",
    "Connect": "127.0.0.1:9090"
  }
}`))
	fmt.Println()
	fmt.Println(d("PIDs and ad-hoc run configs stored in ~/.stunning/pids/"))
}

// --- main --------------------------------------------------------------------

func main() {
	var tf tunnelFlags
	flag.StringVar(&configFile, "config", "tunnels.json", "Config file")
	flag.IntVar(&metricsPort, "metrics-port", 9090, "Metrics HTTP port")
	flag.StringVar(&tf.mode, "mode", "", "Service mode: server | client")
	flag.StringVar(&tf.protocol, "protocol", "", "Tunnel protocol: tcp|udp|tls|h2|ws|udps|dns|icmp|http|https")
	flag.StringVar(&tf.iface, "iface", "", "Local interface: tcp|udp|socks|tun|serial")
	flag.StringVar(&tf.listen, "listen", "", "Listen address")
	flag.StringVar(&tf.connect, "connect", "", "Connect/forward address")
	flag.StringVar(&tf.plugins, "plugins", "", "Plugin chain spec")
	flag.StringVar(&tf.auth, "auth", "", "Auth gate spec")
	flag.StringVar(&tf.knock, "knock", "", "Port-knock gate spec")
	flag.StringVar(&tf.key, "key", "", "TLS key file")
	flag.StringVar(&tf.cert, "cert", "", "TLS cert file")
	flag.StringVar(&tf.device, "device", "", "TUN device name")
	flag.StringVar(&tf.mtu, "mtu", "", "TUN MTU")
	flag.Usage = showHelp
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
			fail("tunnel name required\nUsage: stunning start <name>")
		}
		name := args[1]
		conf, err := resolveConfig(name, tf)
		if err != nil {
			fail("%v", err)
		}
		if err := startTunnel(name, conf, false); err != nil {
			fail("%v", err)
		}

	case "fg":
		if len(args) < 2 {
			fail("tunnel name required\nUsage: stunning fg <name>")
		}
		name := args[1]
		conf, err := resolveConfig(name, tf)
		if err != nil {
			fail("%v", err)
		}
		if err := startTunnel(name, conf, true); err != nil {
			fail("%v", err)
		}

	case "stop":
		if len(args) < 2 {
			fail("tunnel name required\nUsage: stunning stop <name>")
		}
		if err := stopTunnel(args[1]); err != nil {
			fail("%v", err)
		}

	case "status":
		if err := showStatus(configFile); err != nil {
			fail("%v", err)
		}

	case "list":
		if err := listTunnels(configFile); err != nil {
			fail("%v", err)
		}

	case "metrics":
		url := fmt.Sprintf("http://localhost:%d/metrics", metricsPort)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			printErr("could not reach metrics server at %s", url)
			fmt.Println(dimStyle.Render("Make sure a tunnel is running (stunning status)."))
			os.Exit(1)
		}
		defer resp.Body.Close()
		if _, err := os.Stdout.ReadFrom(resp.Body); err != nil {
			fail("reading metrics: %v", err)
		}

	case "version":
		fmt.Println(titleStyle.Render("Stunning") + " v" + version)

	case "help", "-h", "--help":
		showHelp()

	default:
		printErr("unknown command: %s", command)
		fmt.Println(dimStyle.Render("Run 'stunning help' for usage information"))
		os.Exit(1)
	}
}
