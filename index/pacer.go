package index

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// bbsPacer coordinates ALL requests to bbs.robomaster.com:
//
//   - a global minimum interval between requests (a simple slot
//     reservation), capping aggregate QPS regardless of goroutine count
//   - a global cooldown when the forum answers 405 (its rate-limit
//     response): every caller fails fast while the cooldown lasts, so
//     the forum's quota bucket can actually refill. Independent
//     per-goroutine retries were observed to keep the limiter tripped
//     indefinitely at ~1-2 QPS of poke traffic.
//
// The cooldown escalates (30s doubling to 10min) while 405s persist and
// resets to the base after 10 quiet minutes.
var bbsPacer = newPacer()

const (
	pacerBasePenalty = 30 * time.Second
	pacerMaxPenalty  = 10 * time.Minute
	pacerQuietReset  = 10 * time.Minute
)

type pacer struct {
	mu           sync.Mutex
	interval     time.Duration // min spacing between requests
	nextSlot     time.Time     // earliest next request
	blockedUntil time.Time
	penalty      time.Duration
	lastPenalty  time.Time
}

func newPacer() *pacer {
	p := &pacer{
		interval: 25 * time.Millisecond, // ~40 QPS global cap
		penalty:  pacerBasePenalty,
	}
	if v := os.Getenv("RM_SEARCH_BBS_QPS"); v != "" {
		if q, err := strconv.Atoi(v); err == nil && q > 0 {
			p.interval = time.Second / time.Duration(q)
		} else {
			logrus.Warnf("invalid RM_SEARCH_BBS_QPS=%q, using default 40", v)
		}
	}
	return p
}

// reserve blocks for the request's slot. It returns ErrStatusMethodNotAllowed
// immediately while a cooldown is active, so callers reuse their existing
// rate-limit handling without issuing any HTTP traffic.
func (p *pacer) reserve() error {
	p.mu.Lock()
	now := time.Now()

	if p.blockedUntil.After(now) {
		p.mu.Unlock()
		return ErrStatusMethodNotAllowed
	}

	slot := p.nextSlot.Add(p.interval)
	if !slot.After(now) {
		slot = now
	}
	p.nextSlot = slot
	p.mu.Unlock()

	if wait := slot.Sub(now); wait > 0 {
		time.Sleep(wait)
	}
	return nil
}

// penalize is called on a 405 response: it starts or extends the global
// cooldown. Escalation happens on each cooldown expiry that immediately
// sees another 405; a quiet stretch resets to the base penalty.
func (p *pacer) penalize() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.blockedUntil.After(now) {
		// A cooldown is already running; escalate only once it expires
		// and another 405 arrives, so concurrent callers don't stack.
		return
	}
	if now.Sub(p.lastPenalty) > pacerQuietReset {
		p.penalty = pacerBasePenalty
	} else {
		p.penalty *= 2
		if p.penalty > pacerMaxPenalty {
			p.penalty = pacerMaxPenalty
		}
	}
	p.blockedUntil = now.Add(p.penalty)
	p.lastPenalty = now
	logrus.Warnf("forum rate-limited (405); pausing all forum requests for %s", p.penalty)
}
