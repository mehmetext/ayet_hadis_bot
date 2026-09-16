package schedule

import (
	"testing"
	"time"
)

func TestSlotsEvenlyDivideInclusiveWindow(t *testing.T) {
	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		t.Fatal(err)
	}

	slots, err := Slots(time.Date(2026, time.September, 16, 0, 0, 0, 0, location), "06:30", "22:30", 4, location)
	if err != nil {
		t.Fatalf("Slots() error = %v", err)
	}
	want := []string{"06:30", "11:50", "17:10", "22:30"}
	if len(slots) != len(want) {
		t.Fatalf("slot count = %d, want %d", len(slots), len(want))
	}
	for index, slot := range slots {
		if got := slot.Format("15:04"); got != want[index] {
			t.Fatalf("slot %d = %s, want %s", index, got, want[index])
		}
	}
}
