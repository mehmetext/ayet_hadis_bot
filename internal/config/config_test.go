package config

import "testing"

func TestLoadRejectsMissingOperationalConfiguration(t *testing.T) {
	t.Setenv("TIMEZONE", "")
	t.Setenv("SEND_WINDOW_START", "")
	t.Setenv("SEND_WINDOW_END", "")
	t.Setenv("DAILY_NOTIFICATION_COUNT", "")
	t.Setenv("DATA_DIR", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want a missing configuration error")
	}
}

func TestLoadRejectsInvalidNotificationCount(t *testing.T) {
	setValidOperationalConfiguration(t)
	t.Setenv("DAILY_NOTIFICATION_COUNT", "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadRejectsMissingHTTPTimeout(t *testing.T) {
	setValidOperationalConfiguration(t)
	t.Setenv("HTTP_TIMEOUT_SECONDS", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want a missing HTTP timeout error")
	}
}

func TestLoadRejectsMissingConsoleLanguage(t *testing.T) {
	setValidOperationalConfiguration(t)
	t.Setenv("CONSOLE_LANGUAGE", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want a missing console language error")
	}
}

func TestLoadUsesConfiguredValues(t *testing.T) {
	setValidOperationalConfiguration(t)

	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Timezone != "Europe/Istanbul" || configuration.SendWindowStart != "06:30" || configuration.SendWindowEnd != "22:30" || configuration.DailyNotificationCount != 4 || configuration.DataDirectory != "data" {
		t.Fatalf("unexpected configuration: %#v", configuration)
	}
}

func setValidOperationalConfiguration(t *testing.T) {
	t.Helper()
	t.Setenv("TIMEZONE", "Europe/Istanbul")
	t.Setenv("SEND_WINDOW_START", "06:30")
	t.Setenv("SEND_WINDOW_END", "22:30")
	t.Setenv("DAILY_NOTIFICATION_COUNT", "4")
	t.Setenv("HTTP_TIMEOUT_SECONDS", "30")
	t.Setenv("CONSOLE_LANGUAGE", "ara")
	t.Setenv("DATA_DIR", "data")
}
