package control

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daeuniverse/dae/common/consts"
)

type connTransfer struct {
	Upload   atomic.Uint64
	Download atomic.Uint64
}

func (t *connTransfer) recordUpload(n int64) {
	if n > 0 {
		t.Upload.Add(uint64(n))
	}
}

func (t *connTransfer) recordDownload(n int64) {
	if n > 0 {
		t.Download.Add(uint64(n))
	}
}

type transferSnapshot struct {
	Time     time.Time
	Upload   uint64
	Download uint64
}

type bpfTransferSnapshot struct {
	Time     time.Time
	Upload   uint64
	Download uint64
}

type ConnMetadata struct {
	Src           string
	Dst           string
	Domain        string
	Outbound      string
	Dialer        string
	Policy        string
	Network       string
	Pname         string
	Mac           string
	Dscp          uint8
	RuleIndex     int
	OutboundIndex consts.OutboundIndex
	StartTime     time.Time
	State         string
	ClosedAt      time.Time
	ID            string
}

type connMetadataStore struct {
	mu   sync.RWMutex
	data map[net.Conn]ConnMetadata
}

type closedConnStore struct {
	mu   sync.RWMutex
	data []ConnMetadata
}

func newConnMetadataStore() *connMetadataStore {
	return &connMetadataStore{
		data: make(map[net.Conn]ConnMetadata),
	}
}

func newClosedConnStore() *closedConnStore {
	return &closedConnStore{
		data: make([]ConnMetadata, 0),
	}
}

func (s *closedConnStore) Add(meta ConnMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta.State = "closed"
	meta.ClosedAt = time.Now()
	s.data = append(s.data, meta)
}

func (s *closedConnStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	threshold := time.Now().Add(-15 * time.Second)
	kept := s.data[:0]
	for _, m := range s.data {
		if m.ClosedAt.After(threshold) {
			kept = append(kept, m)
		}
	}
	s.data = kept
}

func (s *closedConnStore) Range(f func(meta ConnMetadata) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.data {
		if !f(m) {
			return
		}
	}
}

func (s *closedConnStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *connMetadataStore) Store(conn net.Conn, meta ConnMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta.State = "established"
	s.data[conn] = meta
}

func (s *connMetadataStore) Load(conn net.Conn) (ConnMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meta, ok := s.data[conn]
	return meta, ok
}

func (s *connMetadataStore) Delete(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, conn)
}

func (s *connMetadataStore) Range(f func(conn net.Conn, meta ConnMetadata) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for conn, meta := range s.data {
		if !f(conn, meta) {
			return
		}
	}
}

func (s *connMetadataStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}
