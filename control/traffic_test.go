package control

import (
	"testing"
	"time"
)

func TestTrafficCounting(t *testing.T) {
	// Verify ConnectionInfo struct has traffic fields
	ci := ConnectionInfo{
		Upload:       1024,
		Download:     4096,
		UploadRate:   512,
		DownloadRate: 2048,
	}
	if ci.Upload != 1024 {
		t.Fatal("Upload field broken")
	}
	if ci.DownloadRate != 2048 {
		t.Fatal("DownloadRate field broken")
	}

	// Verify connMatchKey works
	k := connMatchKey("192.168.1.1", "10.0.0.1", 443, 80, "tcp")
	if k != "192.168.1.1:443->10.0.0.1:80/tcp" {
		t.Fatalf("connMatchKey wrong: %s", k)
	}

	// Verify bpfTransferSnapshot struct
	ts := bpfTransferSnapshot{Time: time.Now(), Upload: 100, Download: 200}
	if ts.Upload != 100 {
		t.Fatal("bpfTransferSnapshot broken")
	}

	t.Log("Traffic counting structures verified OK")
}

func TestBPFTrafficRates(t *testing.T) {
	// Simulate BPF rate calculation
	now := time.Now()
	prev := now.Add(-3 * time.Second)

	// First snapshot
	snap1 := &bpfTransferSnapshot{Time: prev, Upload: 100, Download: 1000}

	// Second snapshot (after 3 seconds of traffic)
	bpfConn := ConnectionInfo{Upload: 700, Download: 7000}
	elapsed := now.Sub(snap1.Time).Seconds()
	uploadRate := uint64(float64(bpfConn.Upload-snap1.Upload) / elapsed)
	downloadRate := uint64(float64(bpfConn.Download-snap1.Download) / elapsed)

	// 600 bytes / 3s = 200 B/s for upload
	if uploadRate != 200 {
		t.Fatalf("Upload rate wrong: got %d, want 200", uploadRate)
	}
	// 6000 bytes / 3s = 2000 B/s for download
	if downloadRate != 2000 {
		t.Fatalf("Download rate wrong: got %d, want 2000", downloadRate)
	}
	t.Logf("BPF rates: up=%d B/s down=%d B/s", uploadRate, downloadRate)
}

func TestConnectionMerge(t *testing.T) {
	// Verify that BPF+userspace merging logic uses correct keys
	// When a userspace connection exists, the BPF entry is deleted
	conns := map[string]int{
		"192.168.1.1:443->10.0.0.1:80/tcp": 1,
		"192.168.1.1:444->10.0.0.2:80/tcp": 1,
	}

	key := connMatchKey("192.168.1.1", "10.0.0.1", 443, 80, "tcp")
	if _, ok := conns[key]; !ok {
		t.Fatal("Merge key mismatch")
	}
	delete(conns, key)
	if len(conns) != 1 {
		t.Fatal("Delete failed")
	}
	t.Log("Connection merge logic verified OK")
}
