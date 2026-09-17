package telegram

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
)

func TestWelcomeMessageIncludesScheduleAndSequence(t *testing.T) {
	message := welcomeMessage(WelcomeSettings{Start: "06:30", End: "22:30", DailyCount: 4})
	for _, expected := range []string{"Ayet & Hadis Botu", "06:30", "22:30", "4 bildirim", "ayet ve hadis sırayla", "/start", "/stop", "/language", "/status", "/help"} {
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

func TestCommandDefinitionsExposeSupportedCommands(t *testing.T) {
	commands := commandDefinitions()
	want := map[string]string{
		"start":    "Abone ol ve bildirim dilini seç",
		"stop":     "Bildirimleri durdur",
		"language": "Bildirim dilini değiştir",
		"status":   "Abonelik durumunu gör",
		"help":     "Kullanılabilir komutları göster",
	}

	if len(commands) != len(want) {
		t.Fatalf("command count = %d, want %d", len(commands), len(want))
	}
	for _, command := range commands {
		description, found := want[command.Command]
		if !found {
			t.Fatalf("unexpected command %q", command.Command)
		}
		if command.Description != description {
			t.Fatalf("description for /%s = %q, want %q", command.Command, command.Description, description)
		}
	}
}

func TestAllowedUpdatesOnlyReceiveHandledUpdateTypes(t *testing.T) {
	allowed := allowedUpdates()
	if len(allowed) != 2 || allowed[0] != "message" || allowed[1] != "callback_query" {
		t.Fatalf("allowed updates = %#v, want message and callback_query", allowed)
	}
	_ = bot.AllowedUpdates(allowed)
}

func TestUserUnavailableErrorIncludesForbiddenTelegramErrors(t *testing.T) {
	if !isUserUnavailableError(bot.ErrorForbidden) {
		t.Fatal("forbidden Telegram errors should deactivate the user")
	}
	if isUserUnavailableError(errors.New("temporary network failure")) {
		t.Fatal("temporary errors should remain retryable")
	}
}

func TestIsStartCommandAcceptsDeepLinksButRejectsLongerCommands(t *testing.T) {
	for _, command := range []string{"/start", "/start invite", "/start@ayet_hadis_bot invite"} {
		if !isStartCommand(command) {
			t.Errorf("isStartCommand(%q) = false, want true", command)
		}
	}
	if isStartCommand("/starter") {
		t.Fatal("/starter must not trigger the start flow")
	}
}
