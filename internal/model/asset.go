package model

import "time"

type Asset struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Checksum    string                 `json:"checksum"`
	StoragePath string                 `json:"storage_path"`
	Size        int64                  `json:"size"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UserID      string                 `json:"user_id"`
}

type Version struct {
	AssetID    string                 `json:"asset_id"`
	VersionNum int                    `json:"version_num"`
	Checksum   string                 `json:"checksum"`
	Timestamp  time.Time              `json:"timestamp"`
	Diff       map[string]interface{} `json:"diff"`
}

type AccessControl struct {
	AssetID    string `json:"asset_id"`
	UserID     string `json:"user_id"`
	Permission string `json:"permission"`
}

type SignedURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
