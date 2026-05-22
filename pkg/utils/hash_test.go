package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestSha256Bytes(t *testing.T) {
	got := Sha256Bytes([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("Sha256Bytes(\"hello\") = %q, want %q", got, want)
	}
}

func TestSha256BytesEmpty(t *testing.T) {
	got := Sha256Bytes(nil)
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Fatalf("Sha256Bytes(nil) = %q, want %q", got, want)
	}
}

func TestSha256HexMatchesBytes(t *testing.T) {
	payload := []byte("the quick brown fox jumps over the lazy dog")
	streamed, err := Sha256Hex(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Sha256Hex returned error: %v", err)
	}
	if streamed != Sha256Bytes(payload) {
		t.Fatalf("streamed hash %q != bytes hash %q", streamed, Sha256Bytes(payload))
	}
}

func TestSha256HexReaderError(t *testing.T) {
	_, err := Sha256Hex(errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader, got nil")
	}
}

func TestSha256HexEmptyReader(t *testing.T) {
	got, err := Sha256Hex(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Sha256Hex returned error: %v", err)
	}
	if got != Sha256Bytes(nil) {
		t.Fatalf("empty reader hash %q != empty bytes hash %q", got, Sha256Bytes(nil))
	}
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errBoom
}

var errBoom = &boomError{msg: "boom"}

type boomError struct{ msg string }

func (e *boomError) Error() string { return e.msg }
