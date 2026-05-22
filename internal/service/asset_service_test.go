package service

import (
	"testing"
	"time"

	"github.com/auto-forge-org/asset-cache/internal/cache"
	"github.com/auto-forge-org/asset-cache/internal/storage"
)

func newSvc() *AssetService {
	return NewAssetService(storage.NewMemoryStore(), cache.NewLRU(8), []byte("test-key"))
}

func TestUploadAndGet(t *testing.T) {
	svc := newSvc()
	asset, err := svc.Upload(UploadInput{Name: "logo.png", Type: "image/png", Data: []byte("hello")})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if asset.ID == "" || asset.Checksum == "" {
		t.Fatalf("expected populated asset, got %+v", asset)
	}
	got, data, err := svc.Get(asset.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.ID != asset.ID || string(data) != "hello" {
		t.Fatalf("round-trip mismatch: %+v / %q", got, data)
	}
}

func TestUploadEmptyRejected(t *testing.T) {
	svc := newSvc()
	if _, err := svc.Upload(UploadInput{Name: "x", Data: nil}); err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestUploadDeduplicatesByChecksum(t *testing.T) {
	svc := newSvc()
	a1, _ := svc.Upload(UploadInput{Name: "a.txt", Data: []byte("same")})
	a2, _ := svc.Upload(UploadInput{Name: "b.txt", Data: []byte("same")})
	if a1.ID != a2.ID {
		t.Fatalf("expected dedupe; got %s vs %s", a1.ID, a2.ID)
	}
}

func TestVersioning(t *testing.T) {
	svc := newSvc()
	a, _ := svc.Upload(UploadInput{Name: "f", Data: []byte("v1")})
	v2, err := svc.NewVersion(a.ID, []byte("v2"))
	if err != nil {
		t.Fatalf("new version failed: %v", err)
	}
	if v2.VersionNum != 2 {
		t.Fatalf("expected version 2, got %d", v2.VersionNum)
	}
	versions, _ := svc.Versions(a.ID)
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
}

func TestSearch(t *testing.T) {
	svc := newSvc()
	svc.Upload(UploadInput{Name: "company-logo.png", Data: []byte("a"), Metadata: map[string]interface{}{"tags": []string{"branding", "logo"}}})
	svc.Upload(UploadInput{Name: "invoice.pdf", Data: []byte("b"), Metadata: map[string]interface{}{"tags": []string{"finance"}}})

	got := svc.Search("logo", nil)
	if len(got) != 1 || got[0].Name != "company-logo.png" {
		t.Fatalf("name search failed: %+v", got)
	}
	got = svc.Search("", []string{"branding"})
	if len(got) != 1 {
		t.Fatalf("tag search failed: %+v", got)
	}
	got = svc.Search("", []string{"nonexistent"})
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %+v", got)
	}
}

func TestSignAndVerify(t *testing.T) {
	svc := newSvc()
	a, _ := svc.Upload(UploadInput{Name: "f", Data: []byte("x")})
	signed, err := svc.Sign(a.ID, time.Hour)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if signed.URL == "" || signed.ExpiresAt.IsZero() {
		t.Fatalf("expected signed url, got %+v", signed)
	}
}

func TestSignNotFound(t *testing.T) {
	svc := newSvc()
	if _, err := svc.Sign("missing", time.Hour); err == nil {
		t.Fatal("expected not found")
	}
}

func TestCacheHitTracking(t *testing.T) {
	svc := newSvc()
	a, _ := svc.Upload(UploadInput{Name: "f", Data: []byte("y")})
	_, _, _ = svc.Get(a.ID)
	_, _, _ = svc.Get(a.ID)
	stats := svc.CacheStats()
	if stats.Hits < 1 {
		t.Fatalf("expected at least one cache hit, got %+v", stats)
	}
}
