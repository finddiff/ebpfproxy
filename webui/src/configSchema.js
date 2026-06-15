export const configSchema = {
  global: {
    label: 'Global',
    type: 'flat',
    fields: [
      { name: 'tproxy_port', label: 'TProxy Port', type: 'number', default: 12345, min: 1, max: 65535, desc: 'Transparent proxy port for eBPF traffic interception' },
      { name: 'tproxy_port_protect', label: 'Protect TProxy Port', type: 'boolean', default: true, advanced: true, desc: 'Prevent unsolicited traffic on tproxy port' },
      { name: 'so_mark_from_dae', label: 'SO Mark', type: 'number', default: 0, min: 0, max: 0xFFFFFFFF, advanced: true, desc: 'Socket mark (0=auto)' },
      { name: 'log_level', label: 'Log Level', type: 'enum', default: 'info', options: ['error', 'warn', 'info', 'debug', 'trace'], desc: 'Logging verbosity' },
      { name: 'tcp_check_url', label: 'TCP Check URLs', type: 'string_list', default: 'http://cp.cloudflare.com,1.1.1.1,2606:4700:4700::1111', advanced: true, desc: 'Node TCP connectivity check URLs' },
      { name: 'tcp_check_http_method', label: 'TCP Check Method', type: 'enum', default: 'HEAD', advanced: true, options: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'COPY', 'HEAD', 'OPTIONS', 'LINK', 'UNLINK', 'PURGE', 'LOCK', 'UNLOCK', 'PROPFIND', 'CONNECT', 'TRACE'], desc: 'HTTP method for check URL' },
      { name: 'udp_check_dns', label: 'UDP Check DNS', type: 'string_list', default: 'dns.google:53,8.8.8.8,2001:4860:4860::8888', advanced: true, desc: 'DNS servers for UDP connectivity check' },
      { name: 'check_interval', label: 'Check Interval', type: 'duration', default: '30s', advanced: true, desc: 'Node connectivity check interval' },
      { name: 'check_tolerance', label: 'Check Tolerance', type: 'duration', default: '0', advanced: true, desc: 'Group switch latency tolerance' },
      { name: 'lan_interface', label: 'LAN Interfaces', type: 'string_list', default: '', advanced: true, desc: 'LAN interfaces (kernel >= 5.7)' },
      { name: 'wan_interface', label: 'WAN Interfaces', type: 'string_list', default: 'auto', desc: 'WAN interfaces (use auto to detect)' },
      { name: 'allow_insecure', label: 'Allow Insecure TLS', type: 'boolean', default: false, advanced: true, desc: 'Skip TLS cert verification (not recommended)' },
      { name: 'dial_mode', label: 'Dial Mode', type: 'enum', default: 'domain', options: ['ip', 'domain', 'domain+', 'domain++'], desc: 'Domain routing mode' },
      { name: 'disable_waiting_network', label: 'Disable Wait Network', type: 'boolean', default: false, desc: 'Start without waiting for network' },
      { name: 'auto_config_kernel_parameter', label: 'Auto Config Kernel', type: 'boolean', default: false, desc: 'Auto-configure kernel params (ip_forward, send_redirects)' },
      { name: 'sniffing_timeout', label: 'Sniffing Timeout', type: 'duration', default: '30ms', advanced: true, desc: 'First data sniff timeout (0=disabled)' },
      { name: 'tls_implementation', label: 'TLS Implementation', type: 'enum', default: 'tls', advanced: true, options: ['tls', 'utls'], desc: 'TLS stack implementation' },
      { name: 'utls_imitate', label: 'uTLS Imitate', type: 'enum', default: 'chrome_auto', advanced: true, depends: { field: 'tls_implementation', value: 'utls' },
        options: ['random', 'randomized', 'randomizedalpn', 'randomizednoalpn', 'firefox_auto', 'firefox_55', 'firefox_56', 'firefox_63', 'firefox_65', 'firefox_99', 'firefox_102', 'firefox_105', 'chrome_auto', 'chrome_58', 'chrome_62', 'chrome_70', 'chrome_72', 'chrome_83', 'chrome_87', 'chrome_96', 'chrome_100', 'chrome_102', 'ios_auto', 'ios_11_1', 'ios_12_1', 'ios_13', 'ios_14', 'android_11_okhttp', 'edge_auto', 'edge_85', 'edge_106', 'safari_auto', 'safari_16_0', '360_auto', '360_7_5', '360_11_0', 'qq_auto', 'qq_11_1'],
        desc: 'Browser fingerprint to imitate (utls only)' },
      { name: 'tls_fragment', label: 'TLS Fragment', type: 'boolean', default: false, advanced: true, depends: { field: 'tls_implementation', value: 'utls' }, desc: 'Enable TLS record fragmentation' },
      { name: 'tls_fragment_length', label: 'Fragment Length', type: 'string', default: '50-100', advanced: true, depends: { field: 'tls_fragment', value: true }, desc: 'Fragment length range' },
      { name: 'tls_fragment_interval', label: 'Fragment Interval', type: 'string', default: '10-20', advanced: true, depends: { field: 'tls_fragment', value: true }, desc: 'Fragment interval range (ms)' },
      { name: 'pprof_port', label: 'Pprof Port', type: 'number', default: 0, min: 0, max: 65535, advanced: true, desc: 'Debug pprof port (0=disabled)' },
      { name: 'mptcp', label: 'MPTCP', type: 'boolean', default: false, advanced: true, desc: 'Enable Multipath TCP' },
      { name: 'bootstrap_resolver', label: 'Bootstrap DNS', type: 'string', default: '', advanced: true, desc: 'DNS resolver for bootstrap (IP:port)' },
      { name: 'fallback_resolver', label: 'Fallback DNS', type: 'string', default: '8.8.8.8:53', advanced: true, desc: 'Fallback DNS resolver (IP:port)' },
      { name: 'bandwidth_max_tx', label: 'Max TX Bandwidth', type: 'string', default: '0', advanced: true, desc: 'Max TX bandwidth (0=unlimited)' },
      { name: 'bandwidth_max_rx', label: 'Max RX Bandwidth', type: 'string', default: '0', advanced: true, desc: 'Max RX bandwidth (0=unlimited)' },
      { name: 'udphop_interval', label: 'UDP Hop Interval', type: 'duration', default: '30s', advanced: true, desc: 'UDP hop interval' },
      { name: 'webui_port', label: 'WebUI Port', type: 'number', default: 0, min: 0, max: 65535, desc: 'WebUI HTTP port' },
      { name: 'webui_token', label: 'WebUI Token', type: 'string', default: '', desc: 'WebUI authentication token' },
    ]
  },

  subscription: {
    label: 'Subscription',
    type: 'list',
    itemLabel: 'Subscription URL',
    placeholder: 'https://example.com/sub#tag'
  },

  node: {
    label: 'Node',
    type: 'list',
    itemLabel: 'Node Link',
    placeholder: 'socks5://user:pass@host:port#name'
  },

  group: {
    label: 'Group',
    type: 'group_list',
    policyOptions: [
      { value: 'random', label: 'Random' },
      { value: 'fixed', label: 'Fixed (first node)' },
      { value: 'select', label: 'Select (manual via WebUI)' },
      { value: 'select_fallback', label: 'Select + Fallback' },
      { value: 'min', label: 'Min Last Latency' },
      { value: 'min_avg10', label: 'Min Avg 10 Latencies' },
      { value: 'min_moving_avg', label: 'Min Moving Avg' },
    ],
    filterFunctions: [
      { name: 'name', label: 'Name', desc: 'Match by node name',
        paramKeys: [
          { value: '', label: 'bare (keyword)' },
          { value: 'keyword', label: 'keyword' },
          { value: 'regex', label: 'regex' },
        ] },
      { name: 'subtag', label: 'SubTag', desc: 'Match by subscription tag',
        paramKeys: [
          { value: '', label: 'bare' },
          { value: 'regex', label: 'regex' },
        ] },
      { name: 'link', label: 'Link', desc: 'Match by link (lua script)',
        paramKeys: [
          { value: 'lua', label: 'lua' },
        ] },
    ],
  },

  routing: {
    label: 'Routing',
    type: 'routing',
    functions: [
      { name: 'domain', label: 'Domain', desc: 'Match domain. Supports geosite:category.',
        params: [{ key: '', label: 'Match', placeholder: 'example.com' }],
        paramKeys: [
          { value: '', label: 'bare (suffix)' },
          { value: 'suffix', label: 'suffix' },
          { value: 'keyword', label: 'keyword' },
          { value: 'regex', label: 'regex' },
          { value: 'full', label: 'full' },
          { value: 'geosite', label: 'geosite' },
        ] },
      { name: 'ip', label: 'IP (Dst)', desc: 'Match dest IP. CIDR, geoip:cn.',
        params: [{ key: '', label: 'IP/CIDR', placeholder: '1.1.1.1 or geoip:cn' }] },
      { name: 'dip', label: 'Dst IP', desc: 'Alias for ip.',
        params: [{ key: '', label: 'IP/CIDR', placeholder: '1.1.1.1 or geoip:cn' }] },
      { name: 'sip', label: 'Src IP', desc: 'Match source IP. CIDR.',
        params: [{ key: '', label: 'IP/CIDR', placeholder: '192.168.1.0/24' }] },
      { name: 'sport', label: 'Src Port', desc: 'Match source port. Range: 8000-9000.',
        params: [{ key: '', label: 'Port/Range', placeholder: '8000 or 8000-9000' }] },
      { name: 'dport', label: 'Dst Port', desc: 'Match dest port. Range.',
        params: [{ key: '', label: 'Port/Range', placeholder: '443 or 1-1024' }] },
      { name: 'ipversion', label: 'IP Version', desc: 'Match IP version.',
        params: [{ key: '', label: 'Version', placeholder: '4', options: ['4', '6'] }] },
      { name: 'l4proto', label: 'L4 Protocol', desc: 'Match layer-4 protocol.',
        params: [{ key: '', label: 'Protocol', placeholder: 'tcp', options: ['tcp', 'udp'] }] },
      { name: 'pname', label: 'Process', desc: 'Match process name (WAN mode).',
        params: [{ key: '', label: 'Process', placeholder: 'curl' }] },
      { name: 'mac', label: 'MAC', desc: 'Match source MAC (LAN mode).',
        params: [{ key: '', label: 'MAC', placeholder: '00:11:22:33:44:55' }] },
      { name: 'dscp', label: 'DSCP', desc: 'Match DSCP/ToS field.',
        params: [{ key: '', label: 'DSCP', placeholder: '0' }] },
    ],
    builtinOutbounds: [
      { value: 'direct', label: 'DIRECT' },
      { value: 'block', label: 'BLOCK' },
      { value: 'must_direct', label: 'MUST_DIRECT' },
    ]
  },

  dns: {
    label: 'DNS',
    type: 'mixed',
    fields: [
      { name: 'bind', label: 'DNS Bind', type: 'string', default: 'tcp+udp://0.0.0.0:53', desc: 'DNS server listen address' },
      { name: 'ipversion_prefer', label: 'IP Version Prefer', type: 'number', default: 0, min: 0, max: 6, desc: '0=no preference, 4=prefer IPv4, 6=prefer IPv6' },
      { name: 'optimistic_cache', label: 'Optimistic Cache', type: 'boolean', default: true, advanced: true, desc: 'Enable optimistic DNS caching' },
      { name: 'optimistic_cache_ttl', label: 'Cache TTL', type: 'number', default: 60, min: 0, advanced: true, depends: { field: 'optimistic_cache', value: true }, desc: 'Optimistic cache TTL (seconds)' },
      { name: 'max_cache_size', label: 'Max Cache Size', type: 'number', default: 0, min: 0, advanced: true, desc: 'Max DNS cache entries (0=unlimited)' },
    ],
    subsections: {
      upstream: { label: 'DNS Upstream', keyLabel: 'Name', valLabel: 'URL', placeholder: 'tcp+udp://dns.google:53' },
      hosts: { label: 'DNS Hosts', keyLabel: 'Domain', valLabel: 'IP', placeholder: '192.168.1.1' },
      fixed_domain_ttl: { label: 'Fixed Domain TTL', keyLabel: 'Domain', valLabel: 'TTL (s)', placeholder: '60' },
      dns_routing: { label: 'DNS Routing', type: 'dns_routing' }
    }
  }
}

export function getDefaultState() {
  return {
    global: {},
    dns: {
      upstream: [],
      hosts: [],
      fixed_domain_ttl: [],
      request_rules: [],
      response_rules: [],
      request_fallback: 'asis',
      response_fallback: 'accept',
    },
    subscription: [],
    node: [],
    group: [],
    routing: {
      rules: [],
      fallback: 'direct'
    }
  }
}
