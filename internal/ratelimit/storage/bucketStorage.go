package storage

import (
	"context"
	"ivanjabrony/cloud-test/internal/ratelimit"
	"sync"
)

type BucketStorage struct {
	buckets map[string]*ratelimit.TokenBucket
	mu      sync.RWMutex
}

func NewBucketStorage() *BucketStorage {
	return &BucketStorage{make(map[string]*ratelimit.TokenBucket), sync.RWMutex{}}
}

func (bs *BucketStorage) Store(ctx context.Context, key string, bucket *ratelimit.TokenBucket) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.buckets[key] = bucket
}

func (bs *BucketStorage) Load(ctx context.Context, key string) (bucket *ratelimit.TokenBucket, ok bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	value, ok := bs.buckets[key]
	return value, ok
}
