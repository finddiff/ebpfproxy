package control

import (
	"net"
	"net/netip"
	"time"

	"github.com/cilium/ebpf"
	"github.com/daeuniverse/dae/common/consts"
)

type bpfConnectionEntry struct {
	Src      string
	Dst      string
	SrcPort  uint16
	DstPort  uint16
	L4Proto  uint8
	State    uint8
	Outbound uint8
}

func (c *ControlPlane) readBPFConnections() []ConnectionInfo {
	return c.readBPFConnectionsLimit(1000)
}

func (c *ControlPlane) readBPFConnectionsLimit(maxEntries int) []ConnectionInfo {
	if c == nil || c.core == nil || c.core.bpf == nil {
		return nil
	}

	var connections []ConnectionInfo
	now := time.Now()

	readTcpConnStateMap(c.core.bpf, func(key bpfTuplesKey, state bpfTcpConnState) bool {
		srcIP, dstIP := extractIPsFromKey(key)
		srcPort := ntohs(key.Sport)
		dstPort := ntohs(key.Dport)
		connState := "established"
		switch state.State {
		case 0:
			connState = "syn_sent"
		case 1:
			connState = "established"
		default:
			connState = "closing"
		}
		info := ConnectionInfo{
			Protocol:   "tcp",
			State:      connState,
			SourceIP:   srcIP.String(),
			SourcePort: srcPort,
			DestIP:     dstIP.String(),
			DestPort:   dstPort,
			StartTime:  now,
			Upload:     state.TxBytes,
			Download:   state.RxBytes,
		}
		if state.Meta.Data.HasRouting != 0 {
			info.Outbound = outboundIndexToName(c, state.Meta.Data.Outbound)
			info.Process = ProcessName2String(state.Pname[:])
			info.Dscp = state.Meta.Data.Dscp
			info.Mac = Mac2String(state.Mac[:])
		}
		connections = append(connections, info)
		return len(connections) < maxEntries
	})

	readUdpConnStateMap(c.core.bpf, func(key bpfTuplesKey, state bpfUdpConnState) bool {
		srcIP, dstIP := extractIPsFromKey(key)
		srcPort := ntohs(key.Sport)
		dstPort := ntohs(key.Dport)
		connState := "active"
		info := ConnectionInfo{
			Protocol:   "udp",
			State:      connState,
			SourceIP:   srcIP.String(),
			SourcePort: srcPort,
			DestIP:     dstIP.String(),
			DestPort:   dstPort,
			StartTime:  now,
			Upload:     state.TxBytes,
			Download:   state.RxBytes,
		}
		if state.Meta.Data.HasRouting != 0 {
			info.Outbound = outboundIndexToName(c, state.Meta.Data.Outbound)
			info.Process = ProcessName2String(state.Pname[:])
			info.Dscp = state.Meta.Data.Dscp
			info.Mac = Mac2String(state.Mac[:])
		}
		connections = append(connections, info)
		return len(connections) < maxEntries
	})

	return connections
}

// BPFConnectionCount returns the number of BPF-tracked connections (TCP + UDP).
// This is a lightweight operation that only counts entries, not building full ConnectionInfo objects.
func (c *ControlPlane) BPFConnectionCount() (tcp int, udp int) {
	if c == nil || c.core == nil || c.core.bpf == nil {
		return 0, 0
	}
	readTcpConnStateMap(c.core.bpf, func(_ bpfTuplesKey, _ bpfTcpConnState) bool {
		tcp++
		return true
	})
	readUdpConnStateMap(c.core.bpf, func(_ bpfTuplesKey, _ bpfUdpConnState) bool {
		udp++
		return true
	})
	return
}

func readTcpConnStateMap(bpf *bpfObjects, fn func(key bpfTuplesKey, state bpfTcpConnState) bool) {
	if bpf == nil || bpf.TcpConnStateMap == nil {
		return
	}
	var (
		key   bpfTuplesKey
		state bpfTcpConnState
	)
	iter := bpf.TcpConnStateMap.Iterate()
	for iter.Next(&key, &state) {
		if !fn(key, state) {
			break
		}
	}
}

func readUdpConnStateMap(bpf *bpfObjects, fn func(key bpfTuplesKey, state bpfUdpConnState) bool) {
	if bpf == nil || bpf.UdpConnStateMap == nil {
		return
	}
	var (
		key   bpfTuplesKey
		state bpfUdpConnState
	)
	iter := bpf.UdpConnStateMap.Iterate()
	for iter.Next(&key, &state) {
		if !fn(key, state) {
			break
		}
	}
}

func extractIPsFromKey(key bpfTuplesKey) (srcIP, dstIP netip.Addr) {
	src := netip.AddrFrom16(key.Sip.U6Addr8)
	dst := netip.AddrFrom16(key.Dip.U6Addr8)
	// Unmap IPv4-mapped IPv6
	if src.Is4In6() {
		src = src.Unmap()
	}
	if dst.Is4In6() {
		dst = dst.Unmap()
	}
	return src, dst
}

func ntohs(v uint16) uint16 {
	return (v>>8)&0xff | (v&0xff)<<8
}

func outboundIndexToName(c *ControlPlane, index uint8) string {
	if c == nil {
		return ""
	}
	outboundIdx := consts.OutboundIndex(index)
	names := c.GetOutboundNames()
	idx := int(outboundIdx)
	if idx >= 0 && idx < len(names) && names[idx] != "" {
		return names[idx]
	}
	return ""
}

var _ = ebpf.Map{}
var _ = net.IP{}
