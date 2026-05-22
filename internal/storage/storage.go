package storage

import (
	"errors"
	"sync"

	"github.com/auto-forge-org/asset-cache/internal/model"
)

var ErrNotFound = errors.New("asset not found")

type Store interface {
	Put(asset model.Asset, data []byte) error
	Get(id string) (model.Asset, []byte, error)
	Delete(id string) error
	List() []model.Asset
	FindByChecksum(checksum string) (model.Asset, bool)
	AppendVersion(v model.Version) error
	Versions(assetID string) []model.Version
}

type MemoryStore struct {
	mu       sync.RWMutex
	assets   map[string]model.Asset
	blobs    map[string][]byte
	versions map[string][]model.Version
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		assets:   make(map[string]model.Asset),
		blobs:    make(map[string][]byte),
		versions: make(map[string][]model.Version),
	}
}

func (s *MemoryStore) Put(asset model.Asset, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assets[asset.ID] = asset
	s.blobs[asset.ID] = data
	return nil
}

func (s *MemoryStore) Get(id string) (model.Asset, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assets[id]
	if !ok {
		return model.Asset{}, nil, ErrNotFound
	}
	return a, s.blobs[id], nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[id]; !ok {
		return ErrNotFound
	}
	delete(s.assets, id)
	delete(s.blobs, id)
	delete(s.versions, id)
	return nil
}

func (s *MemoryStore) List() []model.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Asset, 0, len(s.assets))
	for _, a := range s.assets {
		out = append(out, a)
	}
	return out
}

func (s *MemoryStore) FindByChecksum(checksum string) (model.Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.assets {
		if a.Checksum == checksum {
			return a, true
		}
	}
	return model.Asset{}, false
}

func (s *MemoryStore) AppendVersion(v model.Version) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[v.AssetID]; !ok {
		return ErrNotFound
	}
	s.versions[v.AssetID] = append(s.versions[v.AssetID], v)
	return nil
}

func (s *MemoryStore) Versions(assetID string) []model.Version {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Version, len(s.versions[assetID]))
	copy(out, s.versions[assetID])
	return out
}
