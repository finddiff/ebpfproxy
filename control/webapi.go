package control

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/daeuniverse/dae/common/consts"
	"github.com/daeuniverse/dae/component/outbound/dialer"
	"github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/pkg/config_parser"
)

type ProxyGroupInfo struct {
	Name    string            `json:"name"`
	Policy  string            `json:"policy"`
	Servers []ProxyServerInfo `json:"servers"`
}

type ProxyServerInfo struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Protocol  string  `json:"protocol"`
	Latency   float64 `json:"latency_ms"`
	Alive     bool    `json:"alive"`
	Selected  bool    `json:"selected"`
	GroupName string  `json:"group_name"`
}

type RoutingRuleInfo struct {
	Index     int    `json:"index"`
	MatchType string `json:"match_type"`
	Outbound  string `json:"outbound"`
	Not       bool   `json:"not"`
	Must      bool   `json:"must"`
	Mark      uint32 `json:"mark"`
	Rule      string `json:"rule"`
}

type OriginalRuleInfo struct {
	Rule     string
	Outbound string
}

type ConnectionInfo struct {
	ID         string    `json:"id"`
	Protocol   string    `json:"protocol"`
	SourceIP   string    `json:"source_ip"`
	SourcePort uint16    `json:"source_port"`
	DestIP     string    `json:"dest_ip"`
	DestPort   uint16    `json:"dest_port"`
	Outbound   string    `json:"outbound"`
	Dialer     string    `json:"dialer"`
	Domain     string    `json:"domain"`
	Policy     string    `json:"policy"`
	Mac        string    `json:"mac"`
	Process    string    `json:"process"`
	Dscp       uint8     `json:"dscp"`
	Network    string    `json:"network"`
	RuleIndex  int       `json:"rule_index"`
	Duration   float64   `json:"duration_seconds"`
	Upload      uint64    `json:"upload_bytes"`
	Download    uint64    `json:"download_bytes"`
	UploadRate  uint64    `json:"upload_rate"`
	DownloadRate uint64   `json:"download_rate"`
	State      string    `json:"state"`
	StartTime  time.Time `json:"start_time"`
}

func matchTypeToString(mt consts.MatchType) string {
	switch mt {
	case consts.MatchType_DomainSet:
		return "domain"
	case consts.MatchType_IpSet:
		return "ip"
	case consts.MatchType_SourceIpSet:
		return "source_ip"
	case consts.MatchType_Port:
		return "port"
	case consts.MatchType_SourcePort:
		return "source_port"
	case consts.MatchType_L4Proto:
		return "l4proto"
	case consts.MatchType_IpVersion:
		return "ip_version"
	case consts.MatchType_ProcessName:
		return "process_name"
	case consts.MatchType_Dscp:
		return "dscp"
	case consts.MatchType_Mac:
		return "mac"
	case consts.MatchType_Fallback:
		return "fallback"
	default:
		return "unknown"
	}
}

func ruleOutboundToName(f config.FunctionOrString) string {
	switch v := f.(type) {
	case string:
		return v
	case *config_parser.Function:
		if v.Name != "" {
			return v.Name
		}
	case config_parser.Function:
		if v.Name != "" {
			return v.Name
		}
	case *config_parser.RoutingRule:
		return v.Outbound.String(false, true, true)
	}
	return "unknown"
}

func (c *ControlPlane) GetProxyGroups() []ProxyGroupInfo {
	if c == nil || c.outbounds == nil {
		return []ProxyGroupInfo{}
	}

	var groups []ProxyGroupInfo
	for _, dg := range c.outbounds {
		if dg == nil {
			continue
		}

		var servers []ProxyServerInfo
		selectedIdx := dg.GetSelectedIndex()
		for i, d := range dg.Dialers {
			if d == nil {
				continue
			}
			prop := d.Property()
			name := ""
			addr := ""
			link := ""
			if prop != nil {
				name = prop.Name
				addr = prop.Address
				link = prop.Link
			}

			latency := float64(0)
			alive := false

			tcp4 := dialer.NetworkType{
				L4Proto:   consts.L4ProtoStr_TCP,
				IpVersion: consts.IpVersionStr_4,
			}
			if d.MustGetAlive(&tcp4) {
				alive = true
				if l, ok := d.MustGetLatencies10(&tcp4).LastLatency(); ok {
					latency = float64(l.Milliseconds())
				}
			}

			displayName := name
			if displayName == "" {
				displayName = addr
			}
			if displayName == "" {
				displayName = link
			}

			servers = append(servers, ProxyServerInfo{
				Name:      displayName,
				Address:   addr,
				Protocol:  extractProtocolFromLink(link),
				Latency:   latency,
				Alive:     alive,
				Selected:  i == selectedIdx,
				GroupName: dg.Name,
			})
		}

		policy := dg.PolicyString()

		groups = append(groups, ProxyGroupInfo{
			Name:    dg.Name,
			Policy:  policy,
			Servers: servers,
		})
	}
	return groups
}

func extractProtocolFromLink(link string) string {
	for _, proto := range []string{"socks5", "http", "https", "ss", "ssr", "vmess1", "vmess", "vless", "trojan", "tuic", "hysteria2", "juicity"} {
		if len(link) > len(proto) && link[:len(proto)] == proto {
			return proto
		}
	}
	return "unknown"
}

func (c *ControlPlane) GetProxyServers() []ProxyServerInfo {
	groups := c.GetProxyGroups()
	var servers []ProxyServerInfo
	for _, g := range groups {
		servers = append(servers, g.Servers...)
	}
	return servers
}

func (c *ControlPlane) GetRoutingRules() []RoutingRuleInfo {
	// Use original config rules if available (much more accurate than compiled matches)
	if c.originalRules != nil {
		var rules []RoutingRuleInfo
		for i, or := range c.originalRules {
			rules = append(rules, RoutingRuleInfo{
				Index:     i,
				MatchType: "",
				Outbound:  or.Outbound,
				Rule:      or.Rule,
			})
		}
		return rules
	}

	// Fallback to compiled matches
	if c == nil || c.routingMatcher == nil {
		return []RoutingRuleInfo{}
	}
	outboundNames := c.GetOutboundNames()
	var rules []RoutingRuleInfo
	seenRules := make(map[string]bool)
	for _, cm := range c.routingMatcher.compiledMatches {
		if !c.isDisplayRule(cm) { continue }
		outboundName := c.resolveOutboundName(cm, outboundNames)
		ruleDetail := buildRuleDetail(cm)
		dedupKey := ruleDetail + "|" + outboundName
		if seenRules[dedupKey] { continue }
		seenRules[dedupKey] = true
		rules = append(rules, RoutingRuleInfo{
			Index: len(rules), MatchType: matchTypeToString(cm.matchType),
			Outbound: outboundName, Not: cm.not, Must: cm.must, Mark: cm.mark, Rule: ruleDetail,
		})
	}
	return rules
}

func (c *ControlPlane) isDisplayRule(cm compiledRoutingMatch) bool {
	if cm.outbound == consts.OutboundLogicalOr || cm.outbound == consts.OutboundLogicalAnd {
		return false
	}
	if cm.outbound&consts.OutboundLogicalMask == consts.OutboundLogicalMask {
		return false
	}
	return true
}

func (c *ControlPlane) resolveOutboundName(cm compiledRoutingMatch, outboundNames []string) string {
	outboundIdx := int(cm.outbound)
	if outboundIdx == int(consts.OutboundDirect) {
		return "direct"
	}
	if outboundIdx == int(consts.OutboundBlock) {
		return "block"
	}
	if outboundIdx >= 0 && outboundIdx < len(outboundNames) {
		return outboundNames[outboundIdx]
	}
	return "unknown"
}

func (c *ControlPlane) GetDisplayRuleIndex(rawIndex int) int {
	if c == nil || c.routingMatcher == nil || rawIndex < 0 {
		return -1
	}
	if rawIndex >= len(c.routingMatcher.compiledMatches) {
		return -1
	}

	cm := c.routingMatcher.compiledMatches[rawIndex]

	// If original rules are available, match the compiled match back to its source rule
	if c.originalRules != nil {
		outboundName := c.resolveOutboundName(cm, c.GetOutboundNames())
		ruleDetail := buildRuleDetail(cm)

		// Match by rawRule text (exact or prefix) against original rules
		// For geoip/geosite expanded rules, fall back to outbound name matching
		bestIdx := -1
		for i, or := range c.originalRules {
			if or.Outbound == outboundName {
				if bestIdx < 0 {
					bestIdx = i
				}
				// Prefer exact rawRule match over just outbound match
				if cm.rawRule != "" && strings.Contains(or.Rule, cm.rawRule) {
					return i
				}
				_ = ruleDetail
			}
		}
		return bestIdx
	}

	// Fallback: dedup + filter compiled matches
	outboundNames := c.GetOutboundNames()
	seenRules := make(map[string]bool)
	displayIdx := 0
	for i, cm := range c.routingMatcher.compiledMatches {
		if !c.isDisplayRule(cm) { continue }
		outboundName := c.resolveOutboundName(cm, outboundNames)
		ruleDetail := buildRuleDetail(cm)
		dedupKey := ruleDetail + "|" + outboundName
		if seenRules[dedupKey] { continue }
		seenRules[dedupKey] = true
		if i == rawIndex { return displayIdx }
		displayIdx++
	}
	return -1
}

func (c *ControlPlane) GetConnectionRuleIndex(meta ConnMetadata) int {
	return c.getConnectionRuleIndexCached(meta, nil)
}

func (c *ControlPlane) getConnectionRuleIndexCached(meta ConnMetadata, cache map[int]int) int {
	if meta.RuleIndex < 0 {
		return -1
	}
	if cache != nil {
		if v, ok := cache[meta.RuleIndex]; ok {
			return v
		}
	}
	result := c.GetDisplayRuleIndex(meta.RuleIndex)
	if cache != nil && result >= 0 {
		cache[meta.RuleIndex] = result
	}
	return result
}

func buildRuleDetail(cm compiledRoutingMatch) string {
	if cm.rawRule != "" && len(cm.rawRule) <= 100 && !strings.Contains(cm.rawRule, "...") {
		return cm.rawRule
	}
	// For expanded geosite/geoip rules (truncated by Function.String), use short description
	switch cm.matchType {
	case consts.MatchType_ProcessName:
		pname := ProcessName2String(cm.pname[:])
		return "pname(" + pname + ")"
	case consts.MatchType_Port:
		if cm.portStart == cm.portEnd {
			return "dport(" + fmtPortStr(cm.portStart) + ")"
		}
		return "dport(" + fmtPortStr(cm.portStart) + "-" + fmtPortStr(cm.portEnd) + ")"
	case consts.MatchType_SourcePort:
		if cm.portStart == cm.portEnd {
			return "sport(" + fmtPortStr(cm.portStart) + ")"
		}
		return "sport(" + fmtPortStr(cm.portStart) + "-" + fmtPortStr(cm.portEnd) + ")"
	case consts.MatchType_L4Proto:
		switch cm.mask {
		case 6:
			return "l4proto(tcp)"
		case 17:
			return "l4proto(udp)"
		case 1:
			return "l4proto(icmp)"
		default:
			return "l4proto(" + fmtUint32(uint32(cm.mask)) + ")"
		}
	case consts.MatchType_IpVersion:
		switch cm.mask {
		case 4:
			return "ipversion(4)"
		case 6:
			return "ipversion(6)"
		default:
			return "ipversion(" + fmtUint32(uint32(cm.mask)) + ")"
		}
	case consts.MatchType_Dscp:
		return "dscp(" + fmtUint32(uint32(cm.dscp)) + ")"
	case consts.MatchType_IpSet:
		return "dip(ip_ranges)"
	case consts.MatchType_SourceIpSet:
		return "sip(ip_ranges)"
	case consts.MatchType_DomainSet:
		return "domain(domain_set)"
	case consts.MatchType_Mac:
		return "mac(mac_set)"
	case consts.MatchType_Fallback:
		return "fallback"
	default:
		return matchTypeToString(cm.matchType)
	}
}

func fmtPortStr(p uint16) string {
	return fmt.Sprintf("%d", p)
}

func fmtUint32(n uint32) string {
	return fmt.Sprintf("%d", n)
}



func (c *ControlPlane) GetConnections() []ConnectionInfo {
	if c == nil {
		return nil
	}

	var conns []ConnectionInfo
	now := time.Now()
	ruleDisplayCache := make(map[int]int) // cache rawIndex -> displayIndex

	// Layer 1: BPF kernel-tracked connections (syn_sent/established/closing)
	bpfConns := c.readBPFConnections()
	bpfByKey := make(map[string]*ConnectionInfo, len(bpfConns))
	for i := range bpfConns {
		key := connMatchKey(bpfConns[i].SourceIP, bpfConns[i].DestIP, bpfConns[i].SourcePort, bpfConns[i].DestPort, bpfConns[i].Protocol)
		// Calculate BPF-based rates from snapshot
		if last, ok := c.bpfTransferSnapshots.Load(key); ok {
			ls := last.(*bpfTransferSnapshot)
			elapsed := now.Sub(ls.Time).Seconds()
			if elapsed > 0 {
				bpfConns[i].UploadRate = uint64(float64(bpfConns[i].Upload-ls.Upload) / elapsed)
				bpfConns[i].DownloadRate = uint64(float64(bpfConns[i].Download-ls.Download) / elapsed)
			}
		}
		c.bpfTransferSnapshots.Store(key, &bpfTransferSnapshot{
			Time:     now,
			Upload:   bpfConns[i].Upload,
			Download: bpfConns[i].Download,
		})
		bpfByKey[key] = &bpfConns[i]
	}

	// Layer 2: Userspace-tracked connections with rich metadata + transfer rates
	if c.connMetadata != nil {
		c.connMetadata.Range(func(cn net.Conn, meta ConnMetadata) bool {
			info := ConnectionInfo{
				Protocol:  "tcp",
				State:     meta.State,
				StartTime: meta.StartTime,
				Domain:    meta.Domain,
				Outbound:  meta.Outbound,
				Dialer:    meta.Dialer,
				Policy:    meta.Policy,
				Process:   meta.Pname,
				Mac:       meta.Mac,
				Dscp:      meta.Dscp,
				Network:   meta.Network,
				RuleIndex: c.getConnectionRuleIndexCached(meta, ruleDisplayCache),
				ID:        meta.ID,
			}
			if meta.Src != "" {
				if host, _, err := net.SplitHostPort(meta.Src); err == nil {
					info.SourceIP = host
					if ap, err := netip.ParseAddrPort(meta.Src); err == nil {
						info.SourcePort = ap.Port()
					}
				}
			}
			if meta.Dst != "" {
				if host, _, err := net.SplitHostPort(meta.Dst); err == nil {
					info.DestIP = host
					if ap, err := netip.ParseAddrPort(meta.Dst); err == nil {
						info.DestPort = ap.Port()
					}
				}
			}
			info.Duration = now.Sub(meta.StartTime).Seconds()

			// Per-connection transfer rates
			if transfer, ok := c.connTransfers.Load(cn); ok {
				t := transfer.(*connTransfer)
				info.Upload = t.Upload.Load()
				info.Download = t.Download.Load()
				if last, ok := c.lastTransferSnapshots.Load(cn); ok {
					ls := last.(*transferSnapshot)
					elapsed := now.Sub(ls.Time).Seconds()
					if elapsed > 0 {
						info.UploadRate = uint64(float64(info.Upload-ls.Upload) / elapsed)
						info.DownloadRate = uint64(float64(info.Download-ls.Download) / elapsed)
					}
				}
				c.lastTransferSnapshots.Store(cn, &transferSnapshot{
					Time:     now,
					Upload:   info.Upload,
					Download: info.Download,
				})
			}

			// Remove matched BPF entry
			if info.SourceIP != "" {
				key := connMatchKey(info.SourceIP, info.DestIP, info.SourcePort, info.DestPort, info.Protocol)
				delete(bpfByKey, key)
			}
			conns = append(conns, info)
			return true
		})
	} else {
		c.inConnections.Range(func(key, value interface{}) bool {
			conn, ok := key.(net.Conn)
			if !ok { return true }
			info := ConnectionInfo{Protocol: "tcp", State: "established", StartTime: now}
			if remoteAddr := conn.RemoteAddr(); remoteAddr != nil {
				if host, _, err := net.SplitHostPort(remoteAddr.String()); err == nil {
					info.DestIP = host
				}
			}
			if localAddr := conn.LocalAddr(); localAddr != nil {
				if host, _, err := net.SplitHostPort(localAddr.String()); err == nil {
					info.SourceIP = host
				}
			}
			if conn.LocalAddr() != nil && conn.RemoteAddr() != nil {
				info.ID = conn.LocalAddr().String() + "->" + conn.RemoteAddr().String()
			}
			conns = append(conns, info)
			return true
		})
	}

	// Layer 3: Remaining BPF-only connections (no userspace metadata yet)
	// Enrich direct connections with domain from ipDomainMap and rule_index from LRU
	for _, bp := range bpfByKey {
		// Look up domain from IP->domain cache
		if bp.Domain == "" {
			domainKey := fmt.Sprintf("%s:%d", bp.DestIP, bp.DestPort)
			if v, ok := c.ipDomainMap.Load(domainKey); ok {
				bp.Domain = v.(string)
			} else if v, ok := c.ipDomainMap.Load(fmt.Sprintf("%s:0", bp.DestIP)); ok {
				// Fallback: DNS cache stores IP:0 for generic mapping
				bp.Domain = v.(string)
			}
		}
		// Look up rule_index from LRU cache if missing
		if bp.RuleIndex == 0 {
			srcAp, err1 := netip.ParseAddrPort(fmt.Sprintf("%s:%d", bp.SourceIP, bp.SourcePort))
			dstAp, err2 := netip.ParseAddrPort(fmt.Sprintf("%s:%d", bp.DestIP, bp.DestPort))
			if err1 == nil && err2 == nil {
				l4t := consts.L4ProtoType_TCP
				if bp.Protocol == "udp" {
					l4t = consts.L4ProtoType_UDP
				}
				bp.RuleIndex = c.getLRURuleIndex(srcAp, dstAp, l4t)
			}
		}
		// For direct connections, dialer is always "direct" when outbound is direct
		if bp.Outbound == "direct" || bp.Outbound == "" {
			if bp.Dialer == "" {
				bp.Dialer = "direct"
			}
		}
		if bp.Network == "" {
			bp.Network = bp.Protocol
		}
		conns = append(conns, *bp)
	}

	// Layer 4: Closed connections
	if c.closedConns != nil {
		c.closedConns.Range(func(meta ConnMetadata) bool {
			info := ConnectionInfo{
				ID: meta.ID, Protocol: "tcp", SourceIP: meta.Src, DestIP: meta.Dst,
				Domain: meta.Domain, Outbound: meta.Outbound, Dialer: meta.Dialer,
				Policy: meta.Policy, Process: meta.Pname, Mac: meta.Mac, Dscp: meta.Dscp,
				Network: meta.Network, RuleIndex: c.GetConnectionRuleIndex(meta),
				State: "closed", StartTime: meta.StartTime,
				Duration: meta.ClosedAt.Sub(meta.StartTime).Seconds(),
			}
			conns = append(conns, info)
			return true
		})
	}

	return conns
}

func connMatchKey(srcIP, dstIP string, srcPort, dstPort uint16, protocol string) string {
	return fmt.Sprintf("%s:%d->%s:%d/%s", srcIP, srcPort, dstIP, dstPort, protocol)
}

type IncomingConnInfo struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	Protocol   string `json:"protocol"`
}

func (c *ControlPlane) GetIncomingConnections() []IncomingConnInfo {
	if c == nil {
		return nil
	}

	var conns []IncomingConnInfo

	c.inConnections.Range(func(key, value interface{}) bool {
		conn, ok := key.(net.Conn)
		if !ok {
			return true
		}

		info := IncomingConnInfo{
			Protocol: "tcp",
		}
		if conn.LocalAddr() != nil {
			info.LocalAddr = conn.LocalAddr().String()
		}
		if conn.RemoteAddr() != nil {
			info.RemoteAddr = conn.RemoteAddr().String()
		}
		conns = append(conns, info)
		return true
	})

	return conns
}

func (c *ControlPlane) SetSelectedProxy(groupName string, serverIndex int) bool {
	if c == nil || c.outbounds == nil {
		return false
	}
	for _, dg := range c.outbounds {
		if dg == nil || dg.Name != groupName {
			continue
		}
		if serverIndex >= 0 && serverIndex < len(dg.Dialers) {
			dg.SetSelectedIndex(serverIndex)
			return true
		}
	}
	return false
}

func (c *ControlPlane) GetSelectedIndices() map[string]int {
	if c == nil || c.outbounds == nil {
		return nil
	}
	result := make(map[string]int)
	for _, dg := range c.outbounds {
		if dg == nil {
			continue
		}
		result[dg.Name] = dg.GetSelectedIndex()
	}
	return result
}

func (c *ControlPlane) RestoreSelectedIndices(indices map[string]int) {
	if c == nil || c.outbounds == nil || indices == nil {
		return
	}
	for _, dg := range c.outbounds {
		if dg == nil {
			continue
		}
		if idx, ok := indices[dg.Name]; ok {
			dg.SetSelectedIndex(idx)
		}
	}
}

func (c *ControlPlane) CloseConnection(id string) bool {
	if c == nil || c.connMetadata == nil {
		return false
	}
	found := false
	c.connMetadata.Range(func(conn net.Conn, meta ConnMetadata) bool {
		if meta.ID == id {
			if c.closedConns != nil {
				c.closedConns.Add(meta)
			}
			conn.Close()
			found = true
			return false
		}
		return true
	})
	return found
}

func (c *ControlPlane) CloseAllConnections() int {
	if c == nil || c.connMetadata == nil {
		return 0
	}
	count := 0
	c.connMetadata.Range(func(conn net.Conn, meta ConnMetadata) bool {
		if c.closedConns != nil {
			c.closedConns.Add(meta)
		}
		conn.Close()
		count++
		return true
	})
	return count
}

func (c *ControlPlane) GetOutboundNames() []string {
	if c == nil || c.outbounds == nil {
		return nil
	}
	names := make([]string, len(c.outbounds))
	for i, dg := range c.outbounds {
		if dg != nil {
			names[i] = dg.Name
		}
	}
	return names
}