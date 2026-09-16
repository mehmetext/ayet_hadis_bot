package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const defaultContentInterval = 4 * time.Hour

type Config struct {
	ContentInterval  time.Duration
	QuranEdition     string
	HadithEdition    string
	QuranVerseNumber int
	HadithNumber     int
	HTTPTimeout      time.Duration
}

func Load() (Config, error) {
	interval := defaultContentInterval
	if value := os.Getenv("CONTENT_INTERVAL"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("CONTENT_INTERVAL must be a positive duration: %q", value)
		}
		interval = parsed
	}

	timeout := 15 * time.Second
	if value := os.Getenv("HTTP_TIMEOUT_SECONDS"); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds <= 0 {
			return Config{}, fmt.Errorf("HTTP_TIMEOUT_SECONDS must be a positive integer: %q", value)
		}
		timeout = time.Duration(seconds) * time.Second
	}

	return Config{
		ContentInterval:  interval,
		QuranEdition:     envOrDefault("QURAN_EDITION", "quran-uthmani"),
		HadithEdition:    envOrDefault("HADITH_EDITION", "eng-bukhari"),
		QuranVerseNumber: positiveIntOrDefault("QURAN_VERSE_NUMBER", 1),
		HadithNumber:     positiveIntOrDefault("HADITH_NUMBER", 1),
		HTTPTimeout:      timeout,
	}, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func positiveIntOrDefault(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
