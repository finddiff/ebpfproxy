package control

import (
	"net/netip"
	"time"

	"github.com/daeuniverse/dae/common/consts"
	"github.com/maypok86/otter/v2"
)

const (
	routingLRUSize = 8192
	routingLRUTTL  = 30 * time.Minute
)

type routingLRUKey struct {
	SrcIP   [16]byte
	DstIP   [16]byte
	DstPort uint16
	L4Proto uint8
}

type routingLRUValue struct {
	Outbound  consts.OutboundIndex
	Mark      uint32
	Must      bool
	RuleIndex int
}

func newRoutingLRUKey(src, dst netip.AddrPort, l4proto consts.L4ProtoType) routingLRUKey {
	return routingLRUKey{
		SrcIP:   src.Addr().As16(),
		DstIP:   dst.Addr().As16(),
		DstPort: dst.Port(),
		L4Proto: uint8(l4proto),
	}
}

func (c *ControlPlane) initRoutingLRU() {
	cache, err := otter.New[routingLRUKey, routingLRUValue](&otter.Options[routingLRUKey, routingLRUValue]{
		MaximumSize:      routingLRUSize,
		ExpiryCalculator: otter.ExpiryCreating[routingLRUKey, routingLRUValue](routingLRUTTL),
	})
	if err != nil {
		c.log.WithError(err).Warn("Failed to create routing LRU cache")
		return
	}
	c.routingLRU = cache
}

func (c *ControlPlane) routeWithLRU(src, dst netip.AddrPort, l4proto consts.L4ProtoType, doMatch func() (consts.OutboundIndex, uint32, bool, int, error)) (outboundIndex consts.OutboundIndex, mark uint32, must bool, err error) {
	if c.routingLRU == nil {
		outboundIndex, mark, must, _, err = doMatch()
		return
	}

	key := newRoutingLRUKey(src, dst, l4proto)

	if val, ok := c.routingLRU.GetIfPresent(key); ok {
		return val.Outbound, val.Mark, val.Must, nil
	}

	var ruleIndex int
	outboundIndex, mark, must, ruleIndex, err = doMatch()
	if err != nil {
		return
	}

	c.routingLRU.Set(key, routingLRUValue{
		Outbound:  outboundIndex,
		Mark:      mark,
		Must:      must,
		RuleIndex: ruleIndex,
	})

	return
}

func (c *ControlPlane) getLRURuleIndex(src, dst netip.AddrPort, l4proto consts.L4ProtoType) int {
	if c.routingLRU == nil {
		return -1
	}
	key := newRoutingLRUKey(src, dst, l4proto)
	if val, ok := c.routingLRU.GetIfPresent(key); ok {
		return val.RuleIndex
	}
	return -1
}

func (c *ControlPlane) purgeRoutingLRU() {
	if c.routingLRU != nil {
		c.routingLRU.InvalidateAll()
	}
}
