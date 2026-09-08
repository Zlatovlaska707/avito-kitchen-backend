package apikey

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const defaultCacheTTL = 30 * time.Second

type cacheEntry struct {
	rec       *Record
	expiresAt time.Time
}

// CachedRepository wraps Repository with an in-memory prefix lookup cache.
type CachedRepository struct {
	repo *Repository
	ttl  time.Duration

	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func NewCachedRepository(repo *Repository, ttl time.Duration) *CachedRepository {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	return &CachedRepository{repo: repo, ttl: ttl, entries: make(map[string]cacheEntry)}
}

func (c *CachedRepository) LookupByPrefix(ctx context.Context, prefix string) (*Record, error) {
	now := time.Now()
	c.mu.RLock()
	if ent, ok := c.entries[prefix]; ok && now.Before(ent.expiresAt) {
		rec := *ent.rec
		c.mu.RUnlock()
		return &rec, nil
	}
	c.mu.RUnlock()

	rec, err := c.repo.LookupByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.entries[prefix] = cacheEntry{rec: rec, expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()
	return rec, nil
}

func (c *CachedRepository) Invalidate(prefix string) {
	c.mu.Lock()
	delete(c.entries, prefix)
	c.mu.Unlock()
}

func (c *CachedRepository) Issue(ctx context.Context, venueID uuid.UUID, prefix, hash string, scopes []string) (*Record, error) {
	return c.repo.Issue(ctx, venueID, prefix, hash, scopes)
}

func (c *CachedRepository) Revoke(ctx context.Context, venueID, keyID uuid.UUID) (string, error) {
	prefix, err := c.repo.Revoke(ctx, venueID, keyID)
	if err != nil {
		return "", err
	}
	c.Invalidate(prefix)
	return prefix, nil
}

func (c *CachedRepository) ListByVenue(ctx context.Context, venueID uuid.UUID) ([]Record, error) {
	return c.repo.ListByVenue(ctx, venueID)
}
