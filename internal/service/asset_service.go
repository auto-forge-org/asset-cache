package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/auto-forge-org/asset-cache/internal/cache"
	"github.com/auto-forge-org/asset-cache/internal/model"
	"github.com/auto-forge-org/asset-cache/internal/storage"
	"github.com/auto-forge-org/asset-cache/pkg/utils"
)

const MaxUploadBytes int64 = 5 * 1024 * 1024 * 1024 // 5 GB

var (
	ErrTooLarge      = errors.New("file exceeds 5GB limit")
	ErrEmptyFile     = errors.New("file is empty")
	ErrInvalidParams = errors.New("invalid parameters")
)

type AssetService struct {
	store     storage.Store
	cache     *cache.LRU
	signerKey []byte
	now       func() time.Time
}

func NewAssetService(store storage.Store, c *cache.LRU, signerKey []byte) *AssetService {
	return &AssetService{
		store:     store,
		cache:     c,
		signerKey: signerKey,
		now:       time.Now,
	}
}

type UploadInput struct {
	Name     string
	Type     string
	UserID   string
	Data     []byte
	Metadata map[string]interface{}
}

func (s *AssetService) Upload(in UploadInput) (model.Asset, error) {
	if len(in.Data) == 0 {
		return model.Asset{}, ErrEmptyFile
	}
	if int64(len(in.Data)) > MaxUploadBytes {
		return model.Asset{}, ErrTooLarge
	}
	checksum := utils.Sha256Bytes(in.Data)

	if existing, ok := s.store.FindByChecksum(checksum); ok {
		return existing, nil
	}

	id, err := randomID()
	if err != nil {
		return model.Asset{}, err
	}

	now := s.now().UTC()
	asset := model.Asset{
		ID:          id,
		Name:        in.Name,
		Type:        in.Type,
		Checksum:    checksum,
		StoragePath: fmt.Sprintf("local://%s", id),
		Size:        int64(len(in.Data)),
		Metadata:    in.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      in.UserID,
	}
	if err := s.store.Put(asset, in.Data); err != nil {
		return model.Asset{}, err
	}
	if err := s.store.AppendVersion(model.Version{
		AssetID:    id,
		VersionNum: 1,
		Checksum:   checksum,
		Timestamp:  now,
	}); err != nil {
		return model.Asset{}, err
	}
	s.cache.Put(id, in.Data)
	return asset, nil
}

func (s *AssetService) Get(id string) (model.Asset, []byte, error) {
	if data, ok := s.cache.Get(id); ok {
		asset, _, err := s.store.Get(id)
		if err != nil {
			return model.Asset{}, nil, err
		}
		return asset, data, nil
	}
	asset, data, err := s.store.Get(id)
	if err != nil {
		return model.Asset{}, nil, err
	}
	s.cache.Put(id, data)
	return asset, data, nil
}

func (s *AssetService) NewVersion(id string, data []byte) (model.Version, error) {
	if len(data) == 0 {
		return model.Version{}, ErrEmptyFile
	}
	asset, _, err := s.store.Get(id)
	if err != nil {
		return model.Version{}, err
	}
	checksum := utils.Sha256Bytes(data)
	versions := s.store.Versions(id)
	nextNum := len(versions) + 1
	now := s.now().UTC()
	v := model.Version{
		AssetID:    id,
		VersionNum: nextNum,
		Checksum:   checksum,
		Timestamp:  now,
		Diff:       map[string]interface{}{"previous_checksum": asset.Checksum},
	}
	asset.Checksum = checksum
	asset.UpdatedAt = now
	asset.Size = int64(len(data))
	if err := s.store.Put(asset, data); err != nil {
		return model.Version{}, err
	}
	if err := s.store.AppendVersion(v); err != nil {
		return model.Version{}, err
	}
	s.cache.Put(id, data)
	return v, nil
}

func (s *AssetService) Versions(id string) ([]model.Version, error) {
	if _, _, err := s.store.Get(id); err != nil {
		return nil, err
	}
	return s.store.Versions(id), nil
}

func (s *AssetService) Search(query string, tags []string) []model.Asset {
	q := strings.ToLower(strings.TrimSpace(query))
	out := []model.Asset{}
	for _, a := range s.store.List() {
		if q != "" && !strings.Contains(strings.ToLower(a.Name), q) {
			continue
		}
		if !matchTags(a, tags) {
			continue
		}
		out = append(out, a)
	}
	return out
}

func matchTags(a model.Asset, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	raw, ok := a.Metadata["tags"]
	if !ok {
		return false
	}
	assetTags := toStringSlice(raw)
	for _, want := range tags {
		want = strings.ToLower(strings.TrimSpace(want))
		if want == "" {
			continue
		}
		found := false
		for _, have := range assetTags {
			if strings.ToLower(have) == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return strings.Split(t, ",")
	}
	return nil
}

func (s *AssetService) Sign(id string, ttl time.Duration) (model.SignedURL, error) {
	if _, _, err := s.store.Get(id); err != nil {
		return model.SignedURL{}, err
	}
	if ttl <= 0 {
		return model.SignedURL{}, ErrInvalidParams
	}
	exp := s.now().UTC().Add(ttl)
	payload := fmt.Sprintf("%s:%d", id, exp.Unix())
	mac := hmac.New(sha256.New, s.signerKey)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	url := fmt.Sprintf("/api/v1/assets/%s/download?exp=%d&sig=%s", id, exp.Unix(), sig)
	return model.SignedURL{URL: url, ExpiresAt: exp}, nil
}

func (s *AssetService) VerifySignature(id string, exp int64, sig string) bool {
	if s.now().UTC().Unix() > exp {
		return false
	}
	payload := fmt.Sprintf("%s:%d", id, exp)
	mac := hmac.New(sha256.New, s.signerKey)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func (s *AssetService) CacheStats() cache.Stats {
	return s.cache.Stats()
}

func randomID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "asset-" + hex.EncodeToString(b), nil
}
