package tests

import (
	"testing"

	"golang.org/x/time/rate"
)

func TestLimiter(t *testing.T) {
	limiter := rate.NewLimiter(1, 1)
	if !limiter.Allow() {
		t.Error("expected limiter to allow")
	}
	if limiter.Allow() {
		t.Error("expected limiter to not allow")
	}
}
