package cache

import (
	"strconv"
	"testing"
)

func TestPutAndGet(t *testing.T) {
	c := NewLRU(4)
	c.Put("k1", []byte("v1"))
	got, ok := c.Get("k1")
	if !ok || string(got) != "v1" {
		t.Fatalf("Get = %q / %v, want v1 / true", got, ok)
	}
}

func TestGetMiss(t *testing.T) {
	c := NewLRU(2)
	if _, ok := c.Get("absent"); ok {
		t.Fatal("expected miss for absent key")
	}
	stats := c.Stats()
	if stats.Misses != 1 || stats.Hits != 0 {
		t.Fatalf("Stats = %+v, want 1 miss / 0 hits", stats)
	}
}

func TestEviction(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", []byte("1"))
	c.Put("b", []byte("2"))
	c.Put("c", []byte("3"))

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected 'b' to still be present")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected 'c' to still be present")
	}
}

func TestLRUOrderOnGet(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", []byte("1"))
	c.Put("b", []byte("2"))
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected 'a' to be present")
	}
	c.Put("c", []byte("3")) // should evict 'b', not 'a'
	if _, ok := c.Get("b"); ok {
		t.Fatal("expected 'b' to be evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected 'a' to remain after recent access")
	}
}

func TestPutOverwrite(t *testing.T) {
	c := NewLRU(2)
	c.Put("k", []byte("v1"))
	c.Put("k", []byte("v2"))
	got, ok := c.Get("k")
	if !ok || string(got) != "v2" {
		t.Fatalf("after overwrite, Get = %q / %v, want v2 / true", got, ok)
	}
	if stats := c.Stats(); stats.Size != 1 {
		t.Fatalf("Size = %d, want 1", stats.Size)
	}
}

func TestDelete(t *testing.T) {
	c := NewLRU(2)
	c.Put("k", []byte("v"))
	c.Delete("k")
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after delete")
	}
	c.Delete("missing") // no panic on absent key
}

func TestStats(t *testing.T) {
	c := NewLRU(8)
	for i := 0; i < 4; i++ {
		c.Put(strconv.Itoa(i), []byte{byte(i)})
	}
	c.Get("0")
	c.Get("0")
	c.Get("nope")

	s := c.Stats()
	if s.Hits != 2 || s.Misses != 1 || s.Size != 4 {
		t.Fatalf("Stats = %+v, want hits=2 misses=1 size=4", s)
	}
}

func TestDefaultCapacity(t *testing.T) {
	c := NewLRU(0)
	if c.capacity != 128 {
		t.Fatalf("capacity = %d, want 128 default", c.capacity)
	}
}
