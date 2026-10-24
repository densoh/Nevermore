package cmd

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/objects"
)

func TestParseIds(t *testing.T) {
	for _, id := range []int{-91, -90, -88} {
		objects.Rooms[id] = &objects.Room{RoomId: id}
		defer delete(objects.Rooms, id)
	}

	got, err := parseIds([]string{"-91--90,", "-88"}, "ROOM")
	if err == nil {
		t.Errorf("negative ids parsed as a range: %v", got)
	}

	for _, id := range []int{900001, 900002, 900004} {
		objects.Rooms[id] = &objects.Room{RoomId: id}
		defer delete(objects.Rooms, id)
	}
	got, err = parseIds([]string{"900001-900004"}, "ROOM")
	if err != nil || len(got) != 3 {
		t.Errorf("range across a gap = %v, %v; want the 3 that exist", got, err)
	}
	if _, err = parseIds([]string{"900003"}, "ROOM"); err == nil {
		t.Error("a single missing id was accepted")
	}
	if _, err = parseIds([]string{"900004-900001"}, "ROOM"); err == nil {
		t.Error("a backwards range was accepted")
	}
	if _, err = parseIds([]string{"1-5000"}, "ROOM"); err == nil {
		t.Error("an oversized range was accepted")
	}
}

func TestEditAndCompactIds(t *testing.T) {
	list := editIds([]int{5, 1}, []int{2, 3, 3, 9}, true)
	if got := compactIds(list); got != "1-3, 5, 9" {
		t.Errorf("compactIds = %q", got)
	}
	list = editIds(list, []int{2, 9}, false)
	if got := compactIds(list); got != "1, 3, 5" {
		t.Errorf("after remove = %q", got)
	}
}

func TestParseEventExpiry(t *testing.T) {
	day, err := parseEventExpiry([]string{"2026-10-31"})
	want := time.Date(2026, 10, 31, 23, 59, 59, 0, time.Local)
	if err != nil || !day.Equal(want) {
		t.Errorf("date alone = %v, %v; want end of day %v", day, err, want)
	}
	at, err := parseEventExpiry([]string{"2026-10-31", "18:30"})
	if err != nil || !at.Equal(time.Date(2026, 10, 31, 18, 30, 0, 0, time.Local)) {
		t.Errorf("date and time = %v, %v", at, err)
	}
	if _, err := parseEventExpiry([]string{"31/10/2026"}); err == nil {
		t.Error("a bad date was accepted")
	}
}

func TestSplitFlagArgs(t *testing.T) {
	ids, flag, on, remove, err := splitFlagArgs([]string{"2518", "2519-2525", "WATER", "OFF"})
	if err != nil || len(ids) != 2 || flag != "water" || on || remove {
		t.Errorf("switch off = %v %q on=%v remove=%v err=%v", ids, flag, on, remove, err)
	}
	ids, flag, on, remove, err = splitFlagArgs([]string{"REMOVE", "2518", "WATER"})
	if err != nil || len(ids) != 1 || flag != "water" || !remove {
		t.Errorf("remove = %v %q on=%v remove=%v err=%v", ids, flag, on, remove, err)
	}
	for _, bad := range [][]string{
		{"2518", "WATER"},          // no on/off
		{"2518", "WATER", "MAYBE"}, // bad state
		{"2518", "ACTIVE", "OFF"},  // not a switchable flag
		{"REMOVE", "WATER"},        // no rooms
	} {
		if _, _, _, _, err := splitFlagArgs(bad); err == nil {
			t.Errorf("splitFlagArgs(%v) accepted", bad)
		}
	}
}
