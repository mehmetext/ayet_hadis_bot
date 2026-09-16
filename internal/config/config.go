package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultTimezone               = "Europe/Istanbul"
	defaultWindowStart            = "06:30"
	defaultWindowEnd              = "22:30"
	defaultDailyNotificationCount = 4
	defaultDataDirectory          = "data"
	defaultHTTPTimeout            = 30 * time.Second
)

type Config struct {
	Timezone               string
	SendWindowStart        string
	SendWindowEnd          string
	DailyNotificationCount int
	DataDirectory          string
	HTTPTimeout            time.Duration
}

func (configuration Config) DatabasePath() string {
	return filepath.Join(configuration.DataDirectory, "bot.db")
}

func Load() (Config, error) {
	count, err := positiveInteger("DAILY_NOTIFICATION_COUNT", defaultDailyNotificationCount)
	if err != nil {
		return Config{}, err
	}
	timezone := envOrDefault("TIMEZONE", defaultTimezone)
	if _, err := time.LoadLocation(timezone); err != nil {
		return Config{}, fmt.Errorf("TIMEZONE must be a valid IANA time zone: %q", timezone)
	}
	start, end := envOrDefault("SEND_WINDOW_START", defaultWindowStart), envOrDefault("SEND_WINDOW_END", defaultWindowEnd)
	if !validClock(start) || !validClock(end) || start >= end {
		return Config{}, fmt.Errorf("SEND_WINDOW_START and SEND_WINDOW_END must be valid ordered HH:MM times")
	}
	return Config{Timezone: timezone, SendWindowStart: start, SendWindowEnd: end, DailyNotificationCount: count, DataDirectory: envOrDefault("DATA_DIR", defaultDataDirectory), HTTPTimeout: defaultHTTPTimeout}, nil
}
func positiveInteger(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer: %q", name, value)
	}
	return parsed, nil
}
func validClock(value string) bool { _, err := time.Parse("15:04", value); return err == nil }
func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
