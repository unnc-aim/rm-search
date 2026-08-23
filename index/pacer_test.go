package index

import (
	"testing"
	"time"
)

func TestPacerCooldown(t *testing.T) {
	p := newPacer()
	p.interval = time.Millisecond

	// Unblocked: reserve succeeds.
	if err := p.reserve(); err != nil {
		t.Fatalf("reserve before penalty: %v", err)
	}

	// A 405 starts the base cooldown; reserve now fails fast.
	p.penalize()
	if err := p.reserve(); err != ErrStatusMethodNotAllowed {
		t.Fatalf("reserve during cooldown = %v, want ErrStatusMethodNotAllowed", err)
	}

	// Concurrent penalties while blocked must not stack the cooldown.
	p.penalize()
	p.mu.Lock()
	blocked := p.blockedUntil
	p.mu.Unlock()
	if remain := time.Until(blocked); remain > pacerBasePenalty {
		t.Fatalf("cooldown stacked: %v > base %v", remain, pacerBasePenalty)
	}

	// After the cooldown expires, requests flow again...
	p.mu.Lock()
	p.blockedUntil = time.Now().Add(-time.Millisecond)
	p.mu.Unlock()
	if err := p.reserve(); err != nil {
		t.Fatalf("reserve after expiry: %v", err)
	}

	// ...but another 405 inside the quiet window escalates.
	p.penalize()
	p.mu.Lock()
	penalty := p.penalty
	p.mu.Unlock()
	if penalty != pacerBasePenalty*2 {
		t.Fatalf("penalty = %v, want escalated %v", penalty, pacerBasePenalty*2)
	}

	// A quiet stretch resets the penalty to base.
	p.mu.Lock()
	p.lastPenalty = time.Now().Add(-pacerQuietReset - time.Minute)
	p.blockedUntil = time.Now().Add(-time.Millisecond) // not blocked
	p.mu.Unlock()
	p.penalize()
	p.mu.Lock()
	penalty = p.penalty
	p.mu.Unlock()
	if penalty != pacerBasePenalty {
		t.Fatalf("penalty after quiet = %v, want base %v", penalty, pacerBasePenalty)
	}
}
