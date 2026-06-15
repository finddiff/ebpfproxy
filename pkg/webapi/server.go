package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	gonet "net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/control"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	hub         *Hub
	mux         *http.ServeMux
	httpServer  *http.Server
	ln          gonet.Listener
	controlPlane atomicControlPlane
	confPath    string
	configData  atomicConfig
	logBuffer   *LogBuffer
	startTime   time.Time
	log         *logrus.Logger
	token       string
	onConfigSave func() error
}

type atomicControlPlane struct {
	mu sync.RWMutex
	cp *control.ControlPlane
}

func (a *atomicControlPlane) Get() *control.ControlPlane {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cp
}

func (a *atomicControlPlane) Set(cp *control.ControlPlane) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cp = cp
}

type atomicConfig struct {
	mu   sync.RWMutex
	data []byte
}

func (a *atomicConfig) Get() []byte {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

func (a *atomicConfig) Set(data []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.data = data
}

func NewServer(log *logrus.Logger, cp *control.ControlPlane, confPath string, staticFS fs.FS) *Server {
	hub := newHub()
	go hub.run()

	s := &Server{
		hub:       hub,
		mux:       http.NewServeMux(),
		confPath:  confPath,
		logBuffer: NewLogBuffer(2000),
		startTime: time.Now(),
		log:       log,
	}
	if cp != nil {
		s.controlPlane.Set(cp)
	}

	s.registerRoutes(staticFS)
	return s
}

func (s *Server) SetControlPlane(cp *control.ControlPlane) {
	s.controlPlane.Set(cp)
}

func (s *Server) SetConfigData(data []byte) {
	s.configData.Set(data)
}

func (s *Server) SetConfPath(path string) {
	s.confPath = path
}

func (s *Server) SetToken(token string) {
	s.token = token
}

func (s *Server) SetOnConfigSave(fn func() error) {
	s.onConfigSave = fn
}

func (s *Server) tokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next(w, r)
			return
		}
		if r.Header.Get("X-Token") == s.token {
			next(w, r)
			return
		}
		if r.URL.Query().Get("token") == s.token {
			next(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid token", "require_token": "true"})
	}
}

func (s *Server) registerRoutes(staticFS fs.FS) {
	s.mux.HandleFunc("/api/token", s.handleTokenVerify)
	s.mux.HandleFunc("/api/overview", s.tokenMiddleware(s.handleOverview))
	s.mux.HandleFunc("/api/dhcp", s.tokenMiddleware(s.handleDHCP))
	s.mux.HandleFunc("/api/sensors", s.tokenMiddleware(s.handleSensors))
	s.mux.HandleFunc("/api/proxy", s.tokenMiddleware(s.handleProxy))
	s.mux.HandleFunc("/api/proxy/select", s.tokenMiddleware(s.handleProxySelect))
	s.mux.HandleFunc("/api/rules", s.tokenMiddleware(s.handleRules))
	s.mux.HandleFunc("/api/connections", s.tokenMiddleware(s.handleConnections))
	s.mux.HandleFunc("/api/connections/", s.tokenMiddleware(s.handleConnectionAction))
	s.mux.HandleFunc("/api/config/schema", s.tokenMiddleware(s.handleConfigSchema))
	s.mux.HandleFunc("/api/config/validate", s.tokenMiddleware(s.handleConfigValidate))
	s.mux.HandleFunc("/api/config", s.tokenMiddleware(s.handleConfig))
	s.mux.HandleFunc("/api/logs", s.tokenMiddleware(s.handleLogs))
	s.mux.HandleFunc("/api/dns", s.tokenMiddleware(s.handleDNS))
	s.mux.HandleFunc("/ws", s.tokenMiddleware(s.handleWebSocket))

	if staticFS != nil {
		s.mux.Handle("/", http.FileServer(http.FS(staticFS)))
	} else {
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(defaultIndexHTML))
		})
	}
}

func (s *Server) ListenAndServe(addr string) error {
	var err error
	s.ln, err = gonet.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.httpServer = &http.Server{Handler: s.mux}
	return s.httpServer.Serve(s.ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) Addr() gonet.Addr {
	if s.ln != nil {
		return s.ln.Addr()
	}
	return nil
}

func (s *Server) BroadcastJSON(msgType string, data interface{}) {
	msg := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.hub.broadcast <- jsonData
}

func (s *Server) StartBackgroundUpdates(ctx context.Context) {
	go s.statsPoller(ctx)
}

func (s *Server) statsPoller(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendOverviewUpdate()
		}
	}
}

func (s *Server) sendOverviewUpdate() {
	cp := s.controlPlane.Get()
	var connCount, udpSessions int
	var uploadRate, downloadRate, uploadTotal, downloadTotal uint64

	if cp != nil {
		stats := cp.SnapshotRuntimeStats(60, 60)
		bpfTCP, bpfUDP := cp.BPFConnectionCount()
		connCount = cp.ActiveTCPConnections() + bpfTCP
		udpSessions = control.DefaultUdpEndpointPool.Len() + bpfUDP
		uploadRate = stats.UploadRate
		downloadRate = stats.DownloadRate
		uploadTotal = stats.UploadTotal
		downloadTotal = stats.DownloadTotal
	}

	vmem, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(0, false)
	loadAvg, _ := load.Avg()
	netIO, _ := psnet.IOCounters(false)

	var netBytesSent, netBytesRecv uint64
	if len(netIO) > 0 {
		netBytesSent = netIO[0].BytesSent
		netBytesRecv = netIO[0].BytesRecv
	}

	memUsed := uint64(0)
	memTotal := uint64(0)
	memPercent := float64(0)
	if vmem != nil {
		memUsed = vmem.Used
		memTotal = vmem.Total
		memPercent = vmem.UsedPercent
	}

	cpuPct := float64(0)
	if len(cpuPercent) > 0 {
		cpuPct = cpuPercent[0]
	}

	load1, load5, load15 := float64(0), float64(0), float64(0)
	if loadAvg != nil {
		load1 = loadAvg.Load1
		load5 = loadAvg.Load5
		load15 = loadAvg.Load15
	}

	data := map[string]interface{}{
		"timestamp":     time.Now().UnixMilli(),
		"cpu_percent":   cpuPct,
		"mem_used":      memUsed,
		"mem_total":     memTotal,
		"mem_percent":   memPercent,
		"load_1":        load1,
		"load_5":        load5,
		"load_15":       load15,
		"connections":   connCount,
		"udp_sessions":  udpSessions,
		"upload_rate":   uploadRate,
		"download_rate": downloadRate,
		"upload_total":  uploadTotal,
		"download_total": downloadTotal,
		"net_sent":      netBytesSent,
		"net_recv":      netBytesRecv,
		"uptime":        time.Since(s.startTime).Seconds(),
	}
	s.BroadcastJSON("overview", data)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	cp := s.controlPlane.Get()
	var connCount, udpSessions int
	var uploadRate, downloadRate, uploadTotal, downloadTotal uint64
	var samples []control.RuntimeTrafficSample

	if cp != nil {
		stats := cp.SnapshotRuntimeStats(60, 60)
		bpfTCP, bpfUDP := cp.BPFConnectionCount()
		connCount = cp.ActiveTCPConnections() + bpfTCP
		udpSessions = control.DefaultUdpEndpointPool.Len() + bpfUDP
		uploadRate = stats.UploadRate
		downloadRate = stats.DownloadRate
		uploadTotal = stats.UploadTotal
		downloadTotal = stats.DownloadTotal
		samples = stats.Samples
	}

	vmem, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(0, false)
	loadAvg, _ := load.Avg()
	netIO, _ := psnet.IOCounters(false)

	var netBytesSent, netBytesRecv uint64
	if len(netIO) > 0 {
		netBytesSent = netIO[0].BytesSent
		netBytesRecv = netIO[0].BytesRecv
	}

	memUsed, memTotal, memPercent := uint64(0), uint64(0), float64(0)
	if vmem != nil {
		memUsed = vmem.Used
		memTotal = vmem.Total
		memPercent = vmem.UsedPercent
	}
	cpuPct := float64(0)
	if len(cpuPercent) > 0 {
		cpuPct = cpuPercent[0]
	}
	load1, load5, load15 := float64(0), float64(0), float64(0)
	if loadAvg != nil {
		load1 = loadAvg.Load1
		load5 = loadAvg.Load5
		load15 = loadAvg.Load15
	}

	writeJSON(w, map[string]interface{}{
		"timestamp":      time.Now().UnixMilli(),
		"cpu_percent":    cpuPct,
		"mem_used":       memUsed,
		"mem_total":      memTotal,
		"mem_percent":    memPercent,
		"load_1":         load1,
		"load_5":         load5,
		"load_15":        load15,
		"connections":    connCount,
		"udp_sessions":   udpSessions,
		"upload_rate":    uploadRate,
		"download_rate":  downloadRate,
		"upload_total":   uploadTotal,
		"download_total": downloadTotal,
		"net_sent":       netBytesSent,
		"net_recv":       netBytesRecv,
		"uptime":         time.Since(s.startTime).Seconds(),
		"traffic_samples": samples,
	})
}

func (s *Server) handleDHCP(w http.ResponseWriter, r *http.Request) {
	leases, err := getDHCPLeases()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"leases": []interface{}{},
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, map[string]interface{}{
		"leases": leases,
	})
}

func (s *Server) handleSensors(w http.ResponseWriter, r *http.Request) {
	sensors, err := getSensors()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"sensors": []interface{}{},
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, map[string]interface{}{
		"sensors": sensors,
	})
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	cp := s.controlPlane.Get()
	if cp == nil {
		writeJSON(w, map[string]interface{}{
			"groups":  []interface{}{},
			"servers": []interface{}{},
		})
		return
	}

	groups := cp.GetProxyGroups()
	servers := cp.GetProxyServers()

	writeJSON(w, map[string]interface{}{
		"groups":  groups,
		"servers": servers,
	})
}

func (s *Server) handleProxySelect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Group       string `json:"group"`
		ServerIndex int    `json:"server_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cp := s.controlPlane.Get()
	if cp == nil {
		writeError(w, "no control plane", http.StatusServiceUnavailable)
		return
	}

	if ok := cp.SetSelectedProxy(body.Group, body.ServerIndex); !ok {
		writeError(w, "group not found or index out of range", http.StatusNotFound)
		return
	}

	s.logBuffer.Append("Proxy selected: " + body.Group + " #" + fmt.Sprintf("%d", body.ServerIndex))
	writeJSON(w, map[string]interface{}{"status": "ok"})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	cp := s.controlPlane.Get()
	if cp == nil {
		writeJSON(w, map[string]interface{}{
			"rules": []interface{}{},
		})
		return
	}

	rules := cp.GetRoutingRules()
	writeJSON(w, map[string]interface{}{
		"rules": rules,
	})
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	cp := s.controlPlane.Get()
	if cp == nil {
		writeJSON(w, map[string]interface{}{
			"connections": []interface{}{},
		})
		return
	}

	limit, offset := 0, 0
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	conns := cp.GetConnections()
	total := len(conns)

	if limit > 0 {
		if offset >= len(conns) {
			conns = nil
		} else {
			end := offset + limit
			if end > len(conns) {
				end = len(conns)
			}
			conns = conns[offset:end]
		}
	}

	writeJSON(w, map[string]interface{}{
		"connections": conns,
		"total":       total,
	})
}

func (s *Server) handleConfigSchema(w http.ResponseWriter, r *http.Request) {
	schema := buildConfigSchema()
	writeJSON(w, schema)
}

func (s *Server) handleConfigValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Config == "" {
		writeJSON(w, map[string]interface{}{"valid": false, "error": "empty config"})
		return
	}
	sections, err := config_parser.Parse(body.Config)
	if err != nil {
		writeJSON(w, map[string]interface{}{"valid": false, "error": err.Error()})
		return
	}
	_, err = config.New(sections)
	if err != nil {
		writeJSON(w, map[string]interface{}{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"valid": true})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := s.configData.Get()
		if data == nil && s.confPath != "" {
			var err error
			data, err = readConfigFile(s.confPath)
			if err != nil {
				writeJSON(w, map[string]interface{}{
					"config": "",
					"error":  err.Error(),
				})
				return
			}
			s.configData.Set(data)
		}
		writeJSON(w, map[string]interface{}{
			"config": string(data),
		})
	case http.MethodPut, http.MethodPost:
		var body struct {
			Config string `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if s.confPath != "" {
			if err := writeConfigFile(s.confPath, body.Config); err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		s.configData.Set([]byte(body.Config))
		s.logBuffer.Append("Config saved")
		reloaded := false
		if s.onConfigSave != nil {
			if err := s.onConfigSave(); err == nil {
				reloaded = true
				s.logBuffer.Append("Config reload triggered")
			}
		}
		writeJSON(w, map[string]interface{}{"status": "ok", "reloaded": reloaded})
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	entries := s.logBuffer.GetEntries(filter)
	writeJSON(w, map[string]interface{}{
		"entries": entries,
	})
}

func (s *Server) handleConnectionAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cp := s.controlPlane.Get()
	if cp == nil {
		writeError(w, "no control plane", http.StatusServiceUnavailable)
		return
	}

	// /api/connections/  → close all
	// /api/connections/<id> → close one
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/connections/")
	if trimmed == "" || trimmed == "/" {
		count := cp.CloseAllConnections()
		s.logBuffer.Append(fmt.Sprintf("Closed all connections (%d total)", count))
		writeJSON(w, map[string]interface{}{"status": "ok", "closed": count})
		return
	}

	if cp.CloseConnection(trimmed) {
		s.logBuffer.Append("Closed connection: " + trimmed)
		writeJSON(w, map[string]interface{}{"status": "ok"})
	} else {
		writeError(w, "connection not found", http.StatusNotFound)
	}
}

func (s *Server) handleDNS(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Upstreams     []dnsUpstreamInfo `json:"upstreams"`
		RequestRules  []dnsRoutingEntry `json:"request_rules"`
		ResponseRules []dnsRoutingEntry `json:"response_rules"`
		RawConfig     string            `json:"raw_config"`
		CacheSize     int               `json:"cache_size"`
	}{}

	raw := string(s.configData.Get())
	if raw == "" && s.confPath != "" {
		data, _ := os.ReadFile(s.confPath)
		raw = string(data)
	}

	resp.RawConfig = raw
	resp.Upstreams, resp.RequestRules, resp.ResponseRules = parseDnsConfig(raw)

	writeJSON(w, resp)
}

func (s *Server) handleTokenVerify(w http.ResponseWriter, r *http.Request) {
	type tokenStatus struct {
		Required    bool   `json:"required"`
		Valid       bool   `json:"valid"`
		Configured  bool   `json:"configured"`
	}
	token := r.Header.Get("X-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	status := tokenStatus{
		Configured: s.token != "",
		Required:   s.token != "",
		Valid:      s.token == "" || token == s.token,
	}
	writeJSON(w, status)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WithError(err).Warn("websocket upgrade failed")
		return
	}

	client := &Client{
		hub:  s.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func GetWebuiFS(dir string) fs.FS {
	if dir == "" {
		return nil
	}
	return os.DirFS(dir)
}

func (s *Server) AppendLog(text string) {
	s.logBuffer.Append(text)
	s.BroadcastJSON("log_entry", map[string]interface{}{
		"time": time.Now().Format("15:04:05"),
		"text": text,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readConfigFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func writeConfigFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

type dnsUpstreamInfo = struct {
	Tag    string `json:"tag"`
	Link   string `json:"link"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
}

type dnsRoutingEntry = struct {
	Rule     string `json:"rule"`
	Upstream string `json:"upstream"`
}

func parseDnsConfig(raw string) (upstreams []dnsUpstreamInfo, reqRules []dnsRoutingEntry, respRules []dnsRoutingEntry) {
	lines := strings.Split(raw, "\n")
	inDns := false
	inUpstream := false
	inRequest := false
	inResponse := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if !inDns {
			if strings.HasPrefix(trimmed, "dns") && (strings.Contains(trimmed, "{") || strings.HasSuffix(trimmed, "{")) {
				inDns = true
				braceDepth = 0
			}
			continue
		}

		if strings.Contains(trimmed, "{") {
			braceDepth++
		}
		if strings.Contains(trimmed, "}") {
			braceDepth--
			if braceDepth < 0 {
				return
			}
		}

		if strings.HasPrefix(trimmed, "upstream") && strings.Contains(trimmed, "{") {
			inUpstream = true
			continue
		}
		if strings.HasPrefix(trimmed, "routing") && strings.Contains(trimmed, "{") {
			continue
		}
		if strings.HasPrefix(trimmed, "request") && strings.Contains(trimmed, "{") {
			inRequest = true
			continue
		}
		if strings.HasPrefix(trimmed, "response") && strings.Contains(trimmed, "{") {
			inResponse = true
			continue
		}

		if strings.TrimSpace(trimmed) == "}" {
			if inResponse {
				inResponse = false
			} else if inRequest {
				inRequest = false
			} else if inUpstream {
				inUpstream = false
			}
			continue
		}

		if inUpstream {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				tag := strings.TrimSpace(parts[0])
				link := strings.TrimRight(strings.TrimLeft(strings.TrimSpace(parts[1]), "'\""), "'\"")
				scheme := ""
				host := ""
				if idx := strings.Index(link, "://"); idx > 0 {
					scheme = link[:idx]
					remaining := link[idx+3:]
					if idx2 := strings.Index(remaining, "/"); idx2 > 0 {
						host = remaining[:idx2]
					} else {
						host = remaining
					}
				}
				upstreams = append(upstreams, dnsUpstreamInfo{Tag: tag, Link: link, Scheme: scheme, Host: host})
			}
			continue
		}

		if inRequest {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) >= 1 {
				rule := strings.TrimSpace(parts[0])
				upstream := ""
				if len(parts) == 2 {
					upstream = strings.TrimSpace(parts[1])
				}
				reqRules = append(reqRules, dnsRoutingEntry{Rule: rule, Upstream: upstream})
			}
			continue
		}

		if inResponse {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) >= 1 {
				rule := strings.TrimSpace(parts[0])
				upstream := ""
				if len(parts) == 2 {
					upstream = strings.TrimSpace(parts[1])
				}
				respRules = append(respRules, dnsRoutingEntry{Rule: rule, Upstream: upstream})
			}
			continue
		}
	}

	return
}

const defaultIndexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dae WebUI</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0d1117; color: #c9d1d9; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #58a6ff; margin-bottom: 20px; }
        .card { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 20px; margin-bottom: 20px; }
        .tabs { display: flex; gap: 4px; margin-bottom: 20px; flex-wrap: wrap; }
        .tab { padding: 10px 20px; background: #21262d; border: 1px solid #30363d; border-radius: 6px 6px 0 0; cursor: pointer; transition: all 0.2s; }
        .tab:hover { background: #30363d; }
        .tab.active { background: #1f6feb; border-color: #1f6feb; color: #fff; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
        .stat { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 16px; }
        .stat-label { font-size: 12px; color: #8b949e; text-transform: uppercase; margin-bottom: 4px; }
        .stat-value { font-size: 24px; font-weight: 600; color: #58a6ff; }
        .loading { text-align: center; padding: 40px; color: #8b949e; }
        .connection-indicator { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
        .connection-indicator.connected { background: #3fb950; }
        .connection-indicator.disconnected { background: #f85149; }
    </style>
</head>
<body>
<div class="container">
    <h1>Dae WebUI <span class="connection-indicator disconnected" id="ws-indicator"></span></h1>
    <div class="tabs" id="tabs"></div>
    <div id="content"><div class="loading">Loading...</div></div>
</div>
<script>
const WS_URL = 'ws://' + location.host + '/ws';
let ws;
let activeTab = 'overview';

const tabs = [
    {id:'overview',label:'Overview'},
    {id:'dhcp',label:'DHCP'},
    {id:'sensors',label:'Sensors'},
    {id:'proxy',label:'Proxy'},
    {id:'rules',label:'Rules'},
    {id:'connections',label:'Connections'},
    {id:'config',label:'Config'},
    {id:'logs',label:'Logs'}
];

function connect() {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => document.getElementById('ws-indicator').className = 'connection-indicator connected';
    ws.onclose = () => { document.getElementById('ws-indicator').className = 'connection-indicator disconnected'; setTimeout(connect, 3000); };
    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        if (msg.type === 'overview' && activeTab === 'overview') renderOverview(msg.data);
        if (msg.type === 'log_entry') appendLog(msg.data);
    };
}

function buildTabs() {
    document.getElementById('tabs').innerHTML = tabs.map(t => 
        '<div class="tab' + (t.id===activeTab?' active':'') + '" onclick="switchTab(\''+t.id+'\')">'+t.label+'</div>'
    ).join('');
}

function switchTab(id) {
    activeTab = id;
    buildTabs();
    loadTab(id);
}

function loadTab(id) {
    const c = document.getElementById('content');
    c.innerHTML = '<div class="loading">Loading...</div>';
    fetch('/api/' + id).then(r=>r.json()).then(data => {
        renderTab(id, data);
    }).catch(() => c.innerHTML = '<div class="card">Failed to load data</div>');
}

function renderTab(id, data) {
    const c = document.getElementById('content');
    switch(id) {
        case 'overview': renderOverview(data); break;
        case 'dhcp': c.innerHTML = '<div class="card"><h2>DHCP Leases</h2><pre>'+JSON.stringify(data,null,2)+'</pre></div>'; break;
        case 'sensors': c.innerHTML = '<div class="card"><h2>Sensors</h2><pre>'+JSON.stringify(data,null,2)+'</pre></div>'; break;
        case 'proxy': c.innerHTML = '<div class="card"><h2>Proxy Groups & Servers</h2><pre>'+JSON.stringify(data,null,2)+'</pre></div>'; break;
        case 'rules': c.innerHTML = '<div class="card"><h2>Routing Rules</h2><pre>'+JSON.stringify(data,null,2)+'</pre></div>'; break;
        case 'connections': c.innerHTML = '<div class="card"><h2>Connections</h2><pre>'+JSON.stringify(data,null,2)+'</pre></div>'; break;
        case 'config': renderConfig(data); break;
        case 'logs': renderLogs(data); break;
    }
}

function renderOverview(d) {
    document.getElementById('content').innerHTML = 
        '<div class="card"><h2>Overview</h2>' +
        '<div class="stats-grid">' +
        statBox('CPU', (d.cpu_percent||0).toFixed(1)+'%') +
        statBox('Memory', formatBytes(d.mem_used) + ' / ' + formatBytes(d.mem_total)) +
        statBox('Connections', d.connections || 0) +
        statBox('UDP Sessions', d.udp_sessions || 0) +
        statBox('Upload Rate', formatBytes(d.upload_rate||0)+'/s') +
        statBox('Download Rate', formatBytes(d.download_rate||0)+'/s') +
        statBox('Upload Total', formatBytes(d.upload_total||0)) +
        statBox('Download Total', formatBytes(d.download_total||0)) +
        statBox('Uptime', formatDuration(d.uptime||0)) +
        statBox('Load', (d.load_1||0).toFixed(2)+' / '+(d.load_5||0).toFixed(2)+' / '+(d.load_15||0).toFixed(2)) +
        '</div></div>';
}

function renderConfig(d) {
    const cfg = d.config || '';
    document.getElementById('content').innerHTML = 
        '<div class="card"><h2>Configuration</h2>' +
        '<textarea id="cfg-editor" style="width:100%;height:400px;background:#0d1117;color:#c9d1d9;border:1px solid #30363d;border-radius:6px;padding:12px;font-family:monospace;font-size:13px;resize:vertical;">'+escapeHtml(cfg)+'</textarea>' +
        '<button onclick="saveConfig()" style="margin-top:10px;padding:8px 16px;background:#238636;color:#fff;border:none;border-radius:6px;cursor:pointer;">Save</button></div>';
}

function renderLogs(d) {
    const ent = (d.entries||[]).map(e=>'<div style="padding:4px 0;border-bottom:1px solid #21262d;"><span style="color:#8b949e">'+e.time+'</span> <span>'+escapeHtml(e.text)+'</span></div>').join('');
    document.getElementById('content').innerHTML = 
        '<div class="card"><h2>Logs</h2><input id="log-filter" placeholder="Filter..." style="width:100%;padding:8px;margin-bottom:10px;background:#0d1117;color:#c9d1d9;border:1px solid #30363d;border-radius:6px;" oninput="refreshLogs()">' +
        '<div id="log-entries" style="max-height:500px;overflow-y:auto;">'+ent+'</div></div>';
}

function statBox(label, value) {
    return '<div class="stat"><div class="stat-label">'+label+'</div><div class="stat-value">'+value+'</div></div>';
}

function formatBytes(b) { if(!b||b===0)return'0 B'; const u=['B','KB','MB','GB','TB']; const i=Math.floor(Math.log(b)/Math.log(1024)); return parseFloat((b/Math.pow(1024,i)).toFixed(2))+' '+u[i]; }
function formatDuration(s) { const d=Math.floor(s/86400); const h=Math.floor((s%86400)/3600); const m=Math.floor((s%3600)/60); return d+'d '+h+'h '+m+'m'; }
function escapeHtml(t) { return t.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }
function appendLog(e) { if(activeTab!=='logs')return; document.getElementById('log-entries').innerHTML += '<div style="padding:4px 0;border-bottom:1px solid #21262d;"><span style="color:#8b949e">'+e.time+'</span> <span>'+escapeHtml(e.text)+'</span></div>'; }
function saveConfig() { fetch('/api/config',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({config:document.getElementById('cfg-editor').value})}).then(()=>alert('Saved')).catch(()=>alert('Failed')); }
function refreshLogs() { const f=document.getElementById('log-filter')?.value||''; fetch('/api/logs?filter='+encodeURIComponent(f)).then(r=>r.json()).then(renderLogs).catch(()=>{}); }

buildTabs();
loadTab('overview');
connect();
</script>
</body>
</html>`
