package objects

import "testing"

func TestEffectExtendCapsAtMax(t *testing.T) {
	e := NewEffect("60", 0, 0, func(int) {}, func() {})
	if got := e.Extend(10, 120); got != 10 {
		t.Fatalf("first extend added %d, want 10", got)
	}
	total := 10
	for i := 0; i < 10; i++ {
		total += e.Extend(10, 120)
	}
	if total != 60 {
		t.Fatalf("extensions added %d seconds in total, want 60 (cap 120 on a 60s effect)", total)
	}
	if remaining := e.TimeRemaining(); remaining > 120 || remaining < 119 {
		t.Fatalf("time remaining %.1f, want just under 120", remaining)
	}
	if got := e.Extend(10, 120); got != 0 {
		t.Fatalf("extend past cap added %d, want 0", got)
	}
}
