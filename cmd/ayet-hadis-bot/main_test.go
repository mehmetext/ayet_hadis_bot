package main

import (
	"context"
	"testing"
	"time"
)

func TestWaitForStopsWhenContextCancelled(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	cancel()
	if waitFor(context, time.Hour) {
		t.Fatal("waitFor() = true, want false after cancellation")
	}
}
