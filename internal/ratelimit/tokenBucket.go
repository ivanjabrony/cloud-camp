package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	ratePerSec float64
	available  float64
	lastRefill time.Time
	ticker     *time.Ticker
	stopChan   chan struct{}
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, ratePerSec float64) *TokenBucket {
	tb := &TokenBucket{
		capacity:   capacity,
		ratePerSec: ratePerSec,
		available:  float64(capacity),
		lastRefill: time.Now(),
		stopChan:   make(chan struct{}),
	}

	refillInterval := calculateRefillInterval(ratePerSec)
	tb.ticker = time.NewTicker(refillInterval)

	go tb.startRefiller()
	return tb
}

func calculateRefillInterval(ratePerSec float64) time.Duration {
	// higher refresh rate for high prs
	if ratePerSec >= 10 {
		return time.Millisecond * 100
	} else if ratePerSec >= 1 {
		return time.Second / time.Duration(ratePerSec)
	}
	// for low rps refresh every second
	return time.Second
}

func (tb *TokenBucket) startRefiller() {
	for {
		select {
		case <-tb.ticker.C:
			tb.refill()
		case <-tb.stopChan:
			tb.ticker.Stop()
			return
		}
	}
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	elapsed := time.Since(tb.lastRefill).Seconds()
	tokensToAdd := elapsed * tb.ratePerSec
	tb.available = min(tb.available+tokensToAdd, float64(tb.capacity))
	tb.lastRefill = time.Now()
}

// UpdateConfig updates configuration of a persons bucket
func (tb *TokenBucket) UpdateConfig(newCapacity int, newRatePerSec float64) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 1. recalculate cur tokens
	elapsed := time.Since(tb.lastRefill).Seconds()
	tb.available += elapsed * tb.ratePerSec

	// 2. apply new parameters
	tb.ratePerSec = newRatePerSec
	tb.capacity = newCapacity

	// 3. recalculate tokens with new capacity
	if tb.available > float64(tb.capacity) {
		tb.available = float64(tb.capacity)
	} else if tb.available < 0 {
		tb.available = 0
	}

	// 4. refresh ticker
	newInterval := calculateRefillInterval(newRatePerSec)
	tb.ticker.Reset(newInterval)
	tb.lastRefill = time.Now()
}

// Allow checks if it is possible to make a requests (if there's enough tokens in a bucket)
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.available >= 1 {
		tb.available--
		return true
	}
	return false
}

func (tb *TokenBucket) Stop() {
	close(tb.stopChan)
}
