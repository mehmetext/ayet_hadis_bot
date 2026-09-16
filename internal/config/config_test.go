package config

import "testing"

func TestLoadUsesOperationalDefaults(t *testing.T) {
	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Timezone != "Europe/Istanbul" || configuration.SendWindowStart != "06:30" || configuration.SendWindowEnd != "22:30" || configuration.DailyNotificationCount != 4 || configuration.DataDirectory != "data" {
		t.Fatalf("unexpected defaults: %#v", configuration)
	}
}

func TestLoadRejectsInvalidNotificationCount(t *testing.T) {
	t.Setenv("DAILY_NOTIFICATION_COUNT", "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}
