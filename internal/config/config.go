package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ConsoleLanguage        string
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
	if err := loadEnvironmentFile(".env"); err != nil {
		return Config{}, err
	}

	timezone, err := requiredEnvironmentValue("TIMEZONE")
	if err != nil {
		return Config{}, err
	}
	start, err := requiredEnvironmentValue("SEND_WINDOW_START")
	if err != nil {
		return Config{}, err
	}
	end, err := requiredEnvironmentValue("SEND_WINDOW_END")
	if err != nil {
		return Config{}, err
	}
	dataDirectory, err := requiredEnvironmentValue("DATA_DIR")
	if err != nil {
		return Config{}, err
	}
	consoleLanguage, err := requiredEnvironmentValue("CONSOLE_LANGUAGE")
	if err != nil {
		return Config{}, err
	}
	count, err := positiveInteger("DAILY_NOTIFICATION_COUNT")
	if err != nil {
		return Config{}, err
	}
	timeoutSeconds, err := positiveInteger("HTTP_TIMEOUT_SECONDS")
	if err != nil {
		return Config{}, err
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return Config{}, fmt.Errorf("TIMEZONE must be a valid IANA time zone: %q", timezone)
	}
	if !validClock(start) || !validClock(end) || start >= end {
		return Config{}, fmt.Errorf("SEND_WINDOW_START and SEND_WINDOW_END must be valid ordered HH:MM times")
	}
	return Config{ConsoleLanguage: consoleLanguage, Timezone: timezone, SendWindowStart: start, SendWindowEnd: end, DailyNotificationCount: count, DataDirectory: dataDirectory, HTTPTimeout: time.Duration(timeoutSeconds) * time.Second}, nil
}

func positiveInteger(name string) (int, error) {
	value, err := requiredEnvironmentValue(name)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer: %q", name, value)
	}
	return parsed, nil
}

func validClock(value string) bool { _, err := time.Parse("15:04", value); return err == nil }

func requiredEnvironmentValue(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("%s is required; copy .env.example to .env and set a value", name)
	}
	return value, nil
}

func loadEnvironmentFile(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, found := strings.Cut(line, "=")
		name = strings.TrimSpace(strings.TrimPrefix(name, "export "))
		if !found || name == "" {
			return fmt.Errorf("invalid .env entry on line %d", lineNumber)
		}
		if _, exists := os.LookupEnv(name); exists {
			continue
		}
		if err := os.Setenv(name, strings.Trim(strings.TrimSpace(value), "\"'")); err != nil {
			return fmt.Errorf("set %s from .env: %w", name, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
}
