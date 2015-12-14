package proxy

import (
	log "github.com/Sirupsen/logrus"

	"time"
)

const (
	MAX_BACKOFF_DELAY = 2 * time.Second
)

// Backoff provides stepped delay on each failed attempts.  It is hard coded to
// cap the delay at 2 seconds.
type Backoff struct {
	attempts int64
}

func (b *Backoff) min(x, y time.Duration) time.Duration {
	if x < y {
		return x
	} else {
		return y
	}
}

// Delay marks this attempt failed, increments counter, and sleep for no more
// then 2 seconds in this goroutine
func (b *Backoff) Delay() {
	b.attempts = b.attempts + 1
	delay := b.min(time.Duration(b.attempts)*2*time.Millisecond, MAX_BACKOFF_DELAY)
	log.WithFields(log.Fields{"after": delay, "attempts": b.attempts}).Debug("delay")
	time.Sleep(delay)
}

// Reset clears failed attempt counter
func (b *Backoff) Reset() {
	log.WithFields(log.Fields{"attempts": b.attempts}).Debug("reset")
	b.attempts = 0
}

// Attempts reports current failed attempts
func (b *Backoff) Attempts() int64 {
	return b.attempts
}
