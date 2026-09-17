package telegram

import (
	"strings"
	"testing"
)

func TestWelcomeMessageIncludesScheduleAndSequence(t *testing.T) {
	message := welcomeMessage(WelcomeSettings{Start: "06:30", End: "22:30", DailyCount: 4})
	for _, expected := range []string{"Ayet & Hadis Botu", "06:30", "22:30", "4 bildirim", "ayet ve hadis sırayla"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("welcome message missing %q: %s", expected, message)
		}
	}
}

func TestSplitMessagePreservesAllTextWithinLimit(t *testing.T) {
	message := "birinci satır\n" + string(make([]rune, 20)) + "\nson satır"
	chunks := splitMessage(message, 10)
	if len(chunks) < 2 {
		t.Fatal("expected message to be split")
	}
	joined := ""
	for _, chunk := range chunks {
		if len([]rune(chunk)) > 10 {
			t.Fatalf("chunk exceeds limit: %d", len([]rune(chunk)))
		}
		joined += chunk
	}
	if joined != message {
		t.Fatalf("split message changed content: %q", joined)
	}
}
