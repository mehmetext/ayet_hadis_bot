package schedule

import (
	"fmt"
	"time"
)

func Slots(day time.Time, startText, endText string, count int, location *time.Location) ([]time.Time, error) {
	if count <= 0 {
		return nil, fmt.Errorf("daily notification count must be positive")
	}
	start, err := parseClock(day, startText, location)
	if err != nil {
		return nil, err
	}
	end, err := parseClock(day, endText, location)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, fmt.Errorf("send window end must be after start")
	}
	if count == 1 {
		return []time.Time{start}, nil
	}

	interval := end.Sub(start) / time.Duration(count-1)
	slots := make([]time.Time, count)
	for index := range slots {
		slots[index] = start.Add(interval * time.Duration(index))
	}
	return slots, nil
}

func parseClock(day time.Time, value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("15:04", value, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid clock value %q: %w", value, err)
	}
	return time.Date(day.Year(), day.Month(), day.Day(), parsed.Hour(), parsed.Minute(), 0, 0, location), nil
}
