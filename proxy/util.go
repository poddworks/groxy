package proxy

import (
	"time"
)

const (
	MAX_BACKOFF_DELAY = 2 * time.Second
)

type Backoff int64

func (b Backoff) min(x, y time.Duration) time.Duration {
	if x > y {
		return x
	} else {
		return y
	}
}

func (b Backoff) Delay() {
	b = b + 1
	delay := b.min(time.Duration(b)*2*time.Millisecond, MAX_BACKOFF_DELAY)
	time.Sleep(delay)
}

func (b Backoff) Reset() {
	b = 0
}
