package webapi

import "github.com/daeuniverse/dae/common/consts"

func buildConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"sections": map[string]interface{}{
			"global":        buildGlobalSchema(),
			"dns":           buildDnsSchema(),
			"subscription":  buildSubscriptionSchema(),
			"node":          buildNodeSchema(),
			"group":         buildGroupSchema(),
			"routing":       buildRoutingSchema(),
		},
	}
}

func buildGlobalSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "Global",
		"description": "Global configuration settings",
		"type":        "flat",
		"fields": []map[string]interface{}{
			{
				"name": "tproxy_port", "label": "TProxy Port", "type": "number",
				"default": 12345, "min": 1, "max": 65535,
				"description": "Transparent proxy port used by eBPF programs",
			},
			{
				"name": "tproxy_port_protect", "label": "Protect TProxy Port", "type": "boolean",
				"default": true, "advanced": true,
				"description": "Prevent unsolicited traffic on tproxy port",
			},
			{
				"name": "so_mark_from_dae", "label": "SO Mark from DAE", "type": "number",
				"default": 0, "min": 0, "max": 0xFFFFFFFF, "advanced": true,
				"description": "Socket mark for dae-originated traffic (0=auto)",
			},
			{
				"name": "log_level", "label": "Log Level", "type": "enum",
				"default": "info",
				"enum":   []string{"error", "warn", "info", "debug", "trace"},
				"description": "Logging verbosity level",
			},
			{
				"name": "tcp_check_url", "label": "TCP Check URLs", "type": "string_list",
				"default":     "http://cp.cloudflare.com,1.1.1.1,2606:4700:4700::1111",
				"advanced":    true,
				"description": "URLs used for TCP connectivity checks",
			},
			{
				"name": "tcp_check_http_method", "label": "TCP Check HTTP Method", "type": "enum",
				"default": "HEAD", "advanced": true,
				"enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE", "COPY", "HEAD", "OPTIONS", "LINK", "UNLINK", "PURGE", "LOCK", "UNLOCK", "PROPFIND", "CONNECT", "TRACE"},
				"description": "HTTP method for tcp_check_url",
			},
			{
				"name": "udp_check_dns", "label": "UDP Check DNS", "type": "string_list",
				"default":     "dns.google:53,8.8.8.8,2001:4860:4860::8888",
				"advanced":    true,
				"description": "DNS servers for UDP connectivity checks",
			},
			{
				"name": "check_interval", "label": "Check Interval", "type": "duration",
				"default": "30s", "advanced": true,
				"description": "Node connectivity check interval",
			},
			{
				"name": "check_tolerance", "label": "Check Tolerance", "type": "duration",
				"default": "0", "advanced": true,
				"description": "Group switch latency tolerance",
			},
			{
				"name": "lan_interface", "label": "LAN Interfaces", "type": "string_list",
				"default":     "",
				"advanced":    true,
				"description": "LAN interfaces for proxy (requires kernel >= 5.7)",
			},
			{
				"name": "wan_interface", "label": "WAN Interfaces", "type": "string_list",
				"default":     "auto",
				"description": "WAN interfaces (use 'auto' to detect)",
			},
			{
				"name": "allow_insecure", "label": "Allow Insecure TLS", "type": "boolean",
				"default":     false, "advanced": true,
				"description": "Skip TLS certificate verification (not recommended)",
			},
			{
				"name": "dial_mode", "label": "Dial Mode", "type": "enum",
				"default": "domain",
				"enum":    []string{"ip", "domain", "domain+", "domain++"},
				"description": "Domain routing mode: ip (resolve first) / domain (connect to domain) / domain+ / domain++",
			},
			{
				"name": "disable_waiting_network", "label": "Disable Waiting Network", "type": "boolean",
				"default":     false,
				"description": "Start without waiting for network availability",
			},
			{
				"name": "auto_config_kernel_parameter", "label": "Auto Config Kernel", "type": "boolean",
				"default":     false,
				"description": "Auto-configure Linux kernel parameters (ip_forward, send_redirects)",
			},
			{
				"name": "sniffing_timeout", "label": "Sniffing Timeout", "type": "duration",
				"default": "30ms", "advanced": true,
				"description": "Wait timeout for first data sniffing (0=disabled)",
			},
			{
				"name": "tls_implementation", "label": "TLS Implementation", "type": "enum",
				"default": "tls", "advanced": true,
				"enum":    []string{"tls", "utls"},
				"description": "TLS stack: tls (Go crypto/tls) or utls (uTLS fingerprinting)",
			},
			{
				"name": "utls_imitate", "label": "uTLS Imitate", "type": "enum",
				"default": "chrome_auto", "advanced": true,
				"enum": []string{
					"random", "randomized", "randomizedalpn", "randomizednoalpn",
					"firefox_auto", "firefox_55", "firefox_56", "firefox_63", "firefox_65", "firefox_99", "firefox_102", "firefox_105",
					"chrome_auto", "chrome_58", "chrome_62", "chrome_70", "chrome_72", "chrome_83", "chrome_87", "chrome_96", "chrome_100", "chrome_102",
					"ios_auto", "ios_11_1", "ios_12_1", "ios_13", "ios_14",
					"android_11_okhttp",
					"edge_auto", "edge_85", "edge_106",
					"safari_auto", "safari_16_0",
					"360_auto", "360_7_5", "360_11_0",
					"qq_auto", "qq_11_1",
				},
				"dependencies": []map[string]string{{"field": "tls_implementation", "value": "utls"}},
				"description": "Browser ClientHello fingerprint to imitate (only when tls_implementation=utls)",
			},
			{
				"name": "tls_fragment", "label": "TLS Fragment", "type": "boolean",
				"default": false, "advanced": true,
				"dependencies": []map[string]string{{"field": "tls_implementation", "value": "utls"}},
				"description": "Enable TLS record fragmentation",
			},
			{
				"name": "tls_fragment_length", "label": "TLS Fragment Length", "type": "string",
				"default": "50-100", "advanced": true,
				"dependencies": []map[string]string{{"field": "tls_fragment", "value": "true"}},
				"description": "TLS fragment length range (e.g., 50-100)",
			},
			{
				"name": "tls_fragment_interval", "label": "TLS Fragment Interval", "type": "string",
				"default": "10-20", "advanced": true,
				"dependencies": []map[string]string{{"field": "tls_fragment", "value": "true"}},
				"description": "TLS fragment interval range in ms (e.g., 10-20)",
			},
			{
				"name": "pprof_port", "label": "Pprof Port", "type": "number",
				"default": 0, "min": 0, "max": 65535, "advanced": true,
				"description": "pprof debug port (0=disabled)",
			},
			{
				"name": "mptcp", "label": "MPTCP", "type": "boolean",
				"default": false, "advanced": true,
				"description": "Enable Multipath TCP for nodes",
			},
			{
				"name": "bootstrap_resolver", "label": "Bootstrap DNS Resolver", "type": "string",
				"default":     "", "advanced": true,
				"description": "Explicit DNS resolver for bootstrap resolution (IP:port). Falls back to 119.29.29.29:53 and 223.5.5.5:53",
			},
			{
				"name": "fallback_resolver", "label": "Fallback DNS Resolver", "type": "string",
				"default":     "8.8.8.8:53", "advanced": true,
				"description": "Fallback DNS resolver (IP:port)",
			},
			{
				"name": "bandwidth_max_tx", "label": "Max TX Bandwidth", "type": "string",
				"default": "0", "advanced": true,
				"description": "Max TX bandwidth (0=unlimited)",
			},
			{
				"name": "bandwidth_max_rx", "label": "Max RX Bandwidth", "type": "string",
				"default": "0", "advanced": true,
				"description": "Max RX bandwidth (0=unlimited)",
			},
			{
				"name": "udphop_interval", "label": "UDP Hop Interval", "type": "duration",
				"default": "30s", "advanced": true,
				"description": "UDP hop interval",
			},
			{
				"name": "webui_port", "label": "WebUI Port", "type": "number",
				"default": 0, "min": 0, "max": 65535,
				"description": "WebUI HTTP port",
			},
			{
				"name": "webui_token", "label": "WebUI Token", "type": "string",
				"default":     "",
				"description": "WebUI authentication token",
			},
		},
	}
}

func buildDnsSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "DNS",
		"description": "DNS server and routing configuration",
		"type":        "mixed",
		"fields": []map[string]interface{}{
			{
				"name": "bind", "label": "DNS Bind Address", "type": "string",
				"default": "tcp+udp://0.0.0.0:53",
				"description": "Address the DNS server listens on",
			},
			{
				"name": "ipversion_prefer", "label": "IP Version Preference", "type": "enum",
				"default": "0",
				"enum":    []map[string]interface{}{{"label": "No preference", "value": "0"}, {"label": "Prefer IPv4", "value": "4"}, {"label": "Prefer IPv6", "value": "6"}},
				"description": "Preferred IP version for DNS responses",
			},
			{
				"name": "optimistic_cache", "label": "Optimistic Cache", "type": "boolean",
				"default":     true, "advanced": true,
				"description": "Enable optimistic DNS caching",
			},
			{
				"name": "optimistic_cache_ttl", "label": "Optimistic Cache TTL", "type": "number",
				"default":     60, "min": 0, "advanced": true,
				"dependencies": []map[string]string{{"field": "optimistic_cache", "value": "true"}},
				"description": "Optimistic cache TTL in seconds",
			},
			{
				"name": "max_cache_size", "label": "Max Cache Size", "type": "number",
				"default":     0, "min": 0, "advanced": true,
				"description": "Max DNS cache entries (0=unlimited)",
			},
		},
		"subsections": map[string]interface{}{
			"upstream": map[string]interface{}{
				"label":       "DNS Upstream",
				"description": "Upstream DNS servers (name: scheme://host:port)",
				"type":        "map_list",
				"key_label":   "Server Name",
				"value_label": "DNS URL",
				"value_placeholder": "tcp+udp://dns.google:53",
			},
			"hosts": map[string]interface{}{
				"label":       "DNS Hosts",
				"description": "Static hosts entries (domain: ip)",
				"type":        "map_list",
				"key_label":   "Domain",
				"value_label": "IP Address",
				"value_placeholder": "192.168.1.1",
			},
			"fixed_domain_ttl": map[string]interface{}{
				"label":       "Fixed Domain TTL",
				"description": "Fixed TTL for specific domains (domain: seconds)",
				"type":        "map_list",
				"key_label":   "Domain",
				"value_label": "TTL (seconds)",
				"value_placeholder": "60",
			},
			"dns_routing": map[string]interface{}{
				"label":       "DNS Routing",
				"description": "DNS request and response routing rules",
				"type":        "dns_routing",
			},
		},
	}
}

func buildSubscriptionSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "Subscription",
		"description": "Subscription URLs for remote node lists",
		"type":        "list",
		"item_label":  "Subscription URL",
		"placeholder": "https://example.com/sub#tag",
	}
}

func buildNodeSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "Node",
		"description": "Proxy node links",
		"type":        "list",
		"item_label":  "Node Link",
		"placeholder": "socks5://user:pass@host:port#name",
	}
}

func buildGroupSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "Group",
		"description": "Proxy groups with filter and policy",
		"type":        "group_list",
		"policy_options": []map[string]interface{}{
			{"value": "random", "label": "Random"},
			{"value": "fixed", "label": "Fixed (first node)"},
			{"value": "select", "label": "Select (manual)"},
			{"value": "select_fallback", "label": "Select + Fallback"},
			{"value": "min", "label": "Min Last Latency"},
			{"value": "min_avg10", "label": "Min Avg 10 Latencies"},
			{"value": "min_moving_avg", "label": "Min Moving Avg"},
		},
		"fields": []map[string]interface{}{
			{
				"name": "name", "label": "Group Name", "type": "string",
				"description": "Group name (used as outbound in routing rules)",
			},
			{
				"name": "policy", "label": "Policy", "type": "policy",
				"description": "Node selection policy",
			},
			{
				"name": "filter", "label": "Node Filters", "type": "filter_list",
				"description": "Filter nodes (name/subtag/link)",
			},
			{
				"name": "tcp_check_url", "label": "Override TCP Check URLs", "type": "string_list",
				"advanced": true,
				"description": "Override global tcp_check_url for this group",
			},
			{
				"name": "tcp_check_http_method", "label": "Override Check HTTP Method", "type": "enum",
				"default": "", "advanced": true,
				"enum":    []string{"", "GET", "POST", "PUT", "PATCH", "DELETE", "COPY", "HEAD", "OPTIONS", "LINK", "UNLINK", "PURGE", "LOCK", "UNLOCK", "PROPFIND", "CONNECT", "TRACE"},
				"description": "Override global tcp_check_http_method",
			},
			{
				"name": "udp_check_dns", "label": "Override UDP Check DNS", "type": "string_list",
				"advanced": true,
				"description": "Override global udp_check_dns",
			},
			{
				"name": "check_interval", "label": "Override Check Interval", "type": "duration",
				"default": "", "advanced": true,
				"description": "Override global check_interval",
			},
			{
				"name": "check_tolerance", "label": "Override Check Tolerance", "type": "duration",
				"default": "", "advanced": true,
				"description": "Override global check_tolerance",
			},
		},
	}
}

func buildRoutingSchema() map[string]interface{} {
	return map[string]interface{}{
		"label":       "Routing",
		"description": "Traffic routing rules (match → outbound)",
		"type":        "routing",
		"functions": []map[string]interface{}{
			{"name": "domain", "label": "Domain", "description": "Match by domain name (supports geosite:category)",
				"params": []map[string]interface{}{
					{"name": "suffix", "label": "Suffix", "placeholder": "example.com"},
					{"name": "keyword", "label": "Keyword", "placeholder": "google"},
					{"name": "regex", "label": "Regex", "placeholder": "^.*\\.google\\..*$"},
					{"name": "full", "label": "Full", "placeholder": "www.example.com"},
					{"name": "", "label": "(default=suffix)", "placeholder": "example.com"},
				},
			},
			{"name": "ip", "label": "IP (Destination)", "description": "Match destination IP. Supports CIDR and geoip:cn",
				"params": []map[string]interface{}{
					{"name": "", "label": "IP/CIDR/geoip:xx", "placeholder": "1.1.1.1 or geoip:cn"},
				},
			},
			{"name": "dip", "label": "Dst IP", "description": "Alias for ip. Match destination IP.",
				"params": []map[string]interface{}{
					{"name": "", "label": "IP/CIDR/geoip:xx", "placeholder": "1.1.1.1 or geoip:cn"},
				},
			},
			{"name": "sip", "label": "Src IP", "description": "Match source IP. Supports CIDR.",
				"params": []map[string]interface{}{
					{"name": "", "label": "IP/CIDR", "placeholder": "192.168.1.0/24"},
				},
			},
			{"name": "sport", "label": "Src Port", "description": "Match source port. Range supported.",
				"params": []map[string]interface{}{
					{"name": "", "label": "Port/Range", "placeholder": "8000 or 8000-9000"},
				},
			},
			{"name": "dport", "label": "Dst Port", "description": "Match destination port. Range supported.",
				"params": []map[string]interface{}{
					{"name": "", "label": "Port/Range", "placeholder": "443 or 1-1024"},
				},
			},
			{"name": "ipversion", "label": "IP Version", "description": "Match IP version (4 or 6).",
				"params": []map[string]interface{}{
					{"name": "", "label": "Version", "placeholder": "4"},
				},
			},
			{"name": "l4proto", "label": "L4 Protocol", "description": "Match layer-4 protocol.",
				"params": []map[string]interface{}{
					{"name": "", "label": "Protocol", "placeholder": "tcp"},
				},
			},
			{"name": "pname", "label": "Process Name", "description": "Match process name (WAN mode only).",
				"params": []map[string]interface{}{
					{"name": "", "label": "Process name", "placeholder": "curl"},
				},
			},
			{"name": "mac", "label": "MAC Address", "description": "Match source MAC address (LAN mode).",
				"params": []map[string]interface{}{
					{"name": "", "label": "MAC", "placeholder": "00:11:22:33:44:55"},
				},
			},
			{"name": "dscp", "label": "DSCP", "description": "Match DSCP/ToS field value.",
				"params": []map[string]interface{}{
					{"name": "", "label": "DSCP Value", "placeholder": "0"},
				},
			},
		},
		"builtin_outbounds": []map[string]interface{}{
			{"value": "direct", "label": "DIRECT"},
			{"value": "block", "label": "BLOCK"},
			{"value": "must_direct", "label": "MUST_DIRECT"},
		},
	}
}

// Suppress unused import warnings for consts package
var _ = consts.DialerSelectionPolicy_Random
