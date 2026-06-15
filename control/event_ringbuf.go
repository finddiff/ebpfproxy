package control

import (
	"bytes"
	"encoding/binary"
	"net/netip"

	"github.com/cilium/ebpf/ringbuf"
)

// daeEvent matches struct dae_event in tproxy.c (80 bytes)
type daeEvent struct {
	Timestamp uint64
	Type      uint32
	Pid       uint32
	Pname     [16]uint8
	Outbound  uint8
	L4proto   uint8
	Pad       [2]uint8
	Sip       [4]uint32
	Dip       [4]uint32
	Sport     uint16
	Dport     uint16
}

func (c *ControlPlane) startEventRingbufReader() {
	if c.core == nil || c.core.bpf == nil || c.core.bpf.EventRingbuf == nil {
		return
	}

	rd, err := ringbuf.NewReader(c.core.bpf.EventRingbuf)
	if err != nil {
		c.log.WithError(err).Warn("Failed to create event ringbuf reader")
		return
	}

	go func() {
		defer rd.Close()
		var ev daeEvent
		for {
			record, err := rd.Read()
			if err != nil {
				select {
				case <-c.ctx.Done():
					return
				default:
					continue
				}
			}
			if len(record.RawSample) < 80 {
				continue
			}
			_ = binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &ev)
			// Events are consumed to drain the ring buffer.
			// Enrichment of direct connections happens lazily in GetConnections()
			// via c.ipDomainMap lookups and routing LRU cache.
			_ = ev
		}
	}()

	c.log.Info("Event ringbuf reader started")
}

func ipv6U32ToAddr(addr [4]uint32) netip.Addr {
	var b [16]byte
	binary.LittleEndian.PutUint32(b[0:4], addr[0])
	binary.LittleEndian.PutUint32(b[4:8], addr[1])
	binary.LittleEndian.PutUint32(b[8:12], addr[2])
	binary.LittleEndian.PutUint32(b[12:16], addr[3])
	ip := netip.AddrFrom16(b)
	if ip.Is4In6() {
		ip = ip.Unmap()
	}
	return ip
}
