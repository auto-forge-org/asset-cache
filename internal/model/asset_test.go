package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAssetJSONRoundTrip(t *testing.T) {
	original := Asset{
		ID:          "asset-123",
		Name:        "logo.png",
		Type:        "image",
		Checksum:    "abc123",
		StoragePath: "/blobs/asset-123",
		Size:        2048,
		Metadata:    map[string]interface{}{"tag": "branding"},
		CreatedAt:   time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 22, 10, 5, 0, 0, time.UTC),
		UserID:      "user-1",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Asset
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID || decoded.Name != original.Name || decoded.Size != original.Size {
		t.Fatalf("round-trip mismatch: got %+v want %+v", decoded, original)
	}
	if decoded.Metadata["tag"] != "branding" {
		t.Fatalf("metadata lost: %+v", decoded.Metadata)
	}
}

func TestVersionJSONFields(t *testing.T) {
	v := Version{
		AssetID:    "asset-1",
		VersionNum: 2,
		Checksum:   "deadbeef",
		Timestamp:  time.Unix(1700000000, 0).UTC(),
		Diff:       map[string]interface{}{"name": "renamed.png"},
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if raw["asset_id"] != "asset-1" {
		t.Errorf("expected asset_id=asset-1, got %v", raw["asset_id"])
	}
	if raw["version_num"].(float64) != 2 {
		t.Errorf("expected version_num=2, got %v", raw["version_num"])
	}
}

func TestAccessControlJSONFields(t *testing.T) {
	ac := AccessControl{AssetID: "a", UserID: "u", Permission: "edit"}
	data, err := json.Marshal(ac)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"asset_id":"a","user_id":"u","permission":"edit"}`
	if string(data) != want {
		t.Fatalf("AccessControl JSON = %s, want %s", data, want)
	}
}

func TestSignedURLNotExpired(t *testing.T) {
	s := SignedURL{URL: "https://x/y", ExpiresAt: time.Now().Add(time.Hour)}
	if !s.ExpiresAt.After(time.Now()) {
		t.Fatal("expected ExpiresAt in the future")
	}
}
