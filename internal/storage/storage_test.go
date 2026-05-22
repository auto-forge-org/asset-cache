package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/auto-forge-org/asset-cache/internal/model"
)

func newAsset(id, checksum string) model.Asset {
	return model.Asset{
		ID:        id,
		Name:      id + ".bin",
		Checksum:  checksum,
		CreatedAt: time.Now(),
	}
}

func TestPutAndGet(t *testing.T) {
	s := NewMemoryStore()
	asset := newAsset("a1", "sum1")
	data := []byte("hello")

	if err := s.Put(asset, data); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, gotData, err := s.Get("a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "a1" || string(gotData) != "hello" {
		t.Fatalf("Get returned %+v / %q", got, gotData)
	}
}

func TestGetMissing(t *testing.T) {
	s := NewMemoryStore()
	_, _, err := s.Get("nope")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := NewMemoryStore()
	asset := newAsset("a1", "sum1")
	_ = s.Put(asset, []byte("x"))
	if err := s.Delete("a1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, err := s.Get("a1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if err := s.Delete("a1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestList(t *testing.T) {
	s := NewMemoryStore()
	_ = s.Put(newAsset("a1", "s1"), []byte("1"))
	_ = s.Put(newAsset("a2", "s2"), []byte("2"))
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
}

func TestFindByChecksum(t *testing.T) {
	s := NewMemoryStore()
	_ = s.Put(newAsset("a1", "needle"), []byte("x"))
	got, ok := s.FindByChecksum("needle")
	if !ok || got.ID != "a1" {
		t.Fatalf("FindByChecksum hit = %v / %+v", ok, got)
	}
	if _, ok := s.FindByChecksum("missing"); ok {
		t.Fatal("expected miss for unknown checksum")
	}
}

func TestVersions(t *testing.T) {
	s := NewMemoryStore()
	_ = s.Put(newAsset("a1", "s1"), []byte("v1"))

	v1 := model.Version{AssetID: "a1", VersionNum: 1, Checksum: "s1", Timestamp: time.Now()}
	v2 := model.Version{AssetID: "a1", VersionNum: 2, Checksum: "s2", Timestamp: time.Now()}
	if err := s.AppendVersion(v1); err != nil {
		t.Fatalf("AppendVersion v1: %v", err)
	}
	if err := s.AppendVersion(v2); err != nil {
		t.Fatalf("AppendVersion v2: %v", err)
	}

	vs := s.Versions("a1")
	if len(vs) != 2 || vs[0].VersionNum != 1 || vs[1].VersionNum != 2 {
		t.Fatalf("Versions = %+v", vs)
	}

	if err := s.AppendVersion(model.Version{AssetID: "ghost", VersionNum: 1}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for ghost asset, got %v", err)
	}
}

func TestVersionsCopyIsolated(t *testing.T) {
	s := NewMemoryStore()
	_ = s.Put(newAsset("a1", "s1"), nil)
	_ = s.AppendVersion(model.Version{AssetID: "a1", VersionNum: 1})

	vs := s.Versions("a1")
	vs[0].VersionNum = 999

	again := s.Versions("a1")
	if again[0].VersionNum != 1 {
		t.Fatalf("mutating returned slice affected store: %+v", again)
	}
}
