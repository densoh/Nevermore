package objects

import (
	"testing"
	"time"
)

// Daily charges reset once the last refresh is from before 0:00 UTC today.
func TestRefreshDue(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 30, 0, 0, time.UTC)
	cases := []struct {
		name string
		last time.Time
		want bool
	}{
		{"never refreshed", time.Time{}, true},
		{"one minute before midnight", time.Date(2026, 10, 6, 23, 59, 0, 0, time.UTC), true},
		{"exactly midnight", time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC), false},
		{"earlier today", time.Date(2026, 10, 7, 0, 10, 0, 0, time.UTC), false},
		{"yesterday in a local zone, today in UTC", time.Date(2026, 10, 6, 17, 10, 0, 0, time.FixedZone("PDT", -7*3600)), false},
	}
	for _, c := range cases {
		if got := RefreshDue(c.last, now); got != c.want {
			t.Errorf("%s: RefreshDue = %v, want %v", c.name, got, c.want)
		}
	}
}
