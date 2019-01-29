package ratelimit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bucket struct {
	interval   time.Duration
	capacity   uint64
	tokens     uint64
	lastUpdate time.Time
	mutex      sync.Mutex
}

func Parse(burstRate string) (*Bucket, error) {
	parts := strings.SplitN(burstRate, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("burst rate not expressed as fraction")
	}

	capacity, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid capacity in burst rate: %s", err)
	}

	interval, err := time.ParseDuration(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid interval in burst rate: %s", err)
	}

	return TokenBucket(capacity, interval), nil
}

func TokenBucket(capacity uint64, interval time.Duration) *Bucket {
	return &Bucket{
		interval:   interval,
		tokens:     capacity,
		capacity:   capacity,
		lastUpdate: time.Now(),
		mutex:      sync.Mutex{},
	}
}

func (b *Bucket) TakeAt(now time.Time) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// sanity check
	if now.Before(b.lastUpdate) {
		return false
	}

	// compute number of tokens to refill
	elapsed := now.Sub(b.lastUpdate)
	refill := uint64(elapsed / b.interval)

	// refill tokens (up to capacity)
	if (b.tokens + refill) > b.capacity {
		b.tokens = b.capacity
	} else {
		b.tokens += refill
	}

	// update timestamp
	b.lastUpdate = now

	// return false if no tokens left
	if b.tokens == 0 {
		return false
	}

	// take a token
	b.tokens -= 1
	return true
}

func (b *Bucket) Take() bool {
	return b.TakeAt(time.Now())
}
