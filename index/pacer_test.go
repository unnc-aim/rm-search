package index

import (
	"errors"
	"testing"
	"time"
)

func TestPacerCircuitBreaker(t *testing.T) {
	p := newPacer()
	p.interval = time.Millisecond

	// Closed: paced reserves succeed.
	if err := p.reserve(); err != nil {
		t.Fatalf("reserve closed: %v", err)
	}
	p.complete(true)

	// A 405 opens the circuit for the fixed cooldown; reserves fail fast.
	p.penalize()
	if err := p.reserve(); !errors.Is(err, ErrStatusMethodNotAllowed) {
		t.Fatalf("reserve open = %v, want ErrStatusMethodNotAllowed", err)
	}

	// After the cooldown the circuit is half-open: exactly one probe.
	p.mu.Lock()
	p.blockedUntil = time.Now().Add(-time.Millisecond)
	p.mu.Unlock()
	if err := p.reserve(); err != nil {
		t.Fatalf("probe reserve: %v", err)
	}
	// While the probe is in flight everyone else fails fast.
	if err := p.reserve(); !errors.Is(err, ErrStatusMethodNotAllowed) {
		t.Fatal("second concurrent probe admitted; want fast fail")
	}

	// A failing probe reopens the circuit for another full cooldown.
	p.complete(false)
	p.penalize()
	if err := p.reserve(); !errors.Is(err, ErrStatusMethodNotAllowed) {
		t.Fatal("reserve after failed probe should fail")
	}
	p.mu.Lock()
	cooldown := time.Until(p.blockedUntil)
	p.mu.Unlock()
	if cooldown <= pacerCooldown-time.Second {
		t.Fatalf("cooldown = %v, want a fresh full %v", cooldown, pacerCooldown)
	}

	// A succeeding probe closes the circuit; paced traffic resumes.
	p.mu.Lock()
	p.blockedUntil = time.Now().Add(-time.Millisecond)
	p.mu.Unlock()
	if err := p.reserve(); err != nil {
		t.Fatalf("probe reserve: %v", err)
	}
	p.complete(true)
	if err := p.reserve(); err != nil {
		t.Fatalf("reserve after closed: %v", err)
	}
}
