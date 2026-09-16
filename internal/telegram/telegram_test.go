package telegram

import "testing"

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
