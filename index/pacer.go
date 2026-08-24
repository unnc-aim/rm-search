package index

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// bbsPacer coordinates ALL requests to bbs.robomaster.com as a circuit
// breaker with a fixed cooldown:
//
//   - closed (normal): a global minimum interval between requests caps
//     aggregate QPS regardless of goroutine count
//   - open (rate-limited): every caller fails fast with zero HTTP
//     traffic for a fixed 5 minutes
//   - half-open (probing): after the cooldown exactly ONE request goes
//     out. If it succeeds the circuit closes and full (paced) traffic
//     resumes; if it rate-limits again, the cooldown restarts. This
//     repeats until the forum's quota bucket refills.
var bbsPacer = newPacer()

// pacerCooldown is the fixed wait after a 405 and between probes.
const pacerCooldown = 5 * time.Minute

type pacer struct {
	mu           sync.Mutex
	interval     time.Duration // min spacing between requests (closed)
	nextSlot     time.Time     // earliest next request (closed)
	blockedUntil time.Time     // open until this instant
	wasBlocked   bool          // a cooldown has opened; half-open until a success
	probing      bool          // half-open: the single in-flight probe
}

func newPacer() *pacer {
	p := &pacer{
		interval: 25 * time.Millisecond, // ~40 QPS global cap
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

// reserve admits one request. It returns ErrStatusMethodNotAllowed
// immediately while the circuit is open or while another probe is in
// flight, so callers reuse their existing rate-limit handling without
// issuing any HTTP traffic.
func (p *pacer) reserve() error {
	p.mu.Lock()
	now := time.Now()

	if p.blockedUntil.After(now) {
		p.mu.Unlock()
		return ErrStatusMethodNotAllowed
	}

	if p.wasBlocked {
		// Half-open: allow exactly one probe at a time.
		if p.probing {
			p.mu.Unlock()
			return ErrStatusMethodNotAllowed
		}
		p.probing = true
		p.mu.Unlock()
		return nil
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

// complete reports the outcome of a request admitted by reserve.
// A success closes a half-open circuit; any outcome releases the probe
// slot so the next single probe can go out.
func (p *pacer) complete(ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.probing = false
	if ok {
		p.wasBlocked = false
	}
}

// penalize is called on a 405 response: (re)open the circuit for the
// fixed cooldown.
func (p *pacer) penalize() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.probing = false
	p.wasBlocked = true
	p.blockedUntil = time.Now().Add(pacerCooldown)
	logrus.Warnf("forum rate-limited (405); pausing all forum requests for %s", pacerCooldown)
}
