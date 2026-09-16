package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadUsesConfigurableInterval(t *testing.T) {
	t.Setenv("CONTENT_INTERVAL", "2h30m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ContentInterval != 2*time.Hour+30*time.Minute {
		t.Fatalf("interval = %s", cfg.ContentInterval)
	}

	_ = os.Unsetenv("CONTENT_INTERVAL")
}
