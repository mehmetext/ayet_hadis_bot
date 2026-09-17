package delivery

import (
	"strings"
	"testing"

	"github.com/mehmetext/ayet-hadis-bot/internal/catalog"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/quranenc"
)

func TestFormatVerseEscapesTelegramText(t *testing.T) {
	formatted := formatVerse(catalog.Language{Code: "tur", Attribution: "QuranEnc"}, quranenc.Verse{
		SurahNumber: 1, AyahNumber: 1, ArabicText: "<Ayet>", Translation: "<Çeviri>",
	})

	if !strings.Contains(formatted.Telegram, "&lt;Ayet&gt;") || !strings.Contains(formatted.Telegram, "&lt;Çeviri&gt;") {
		t.Fatalf("telegram message did not escape text: %q", formatted.Telegram)
	}
}

func TestFormatHadithIncludesReferenceAndGrade(t *testing.T) {
	formatted := formatHadith(catalog.Language{Attribution: "HadeethEnc"}, hadeethenc.Hadith{
		Reference: "Buhari 1", Hadeeth: "Hadis metni", Grade: "Sahih",
	})

	for _, expected := range []string{"Buhari 1", "Hadis metni", "Sahih"} {
		if !strings.Contains(formatted.Telegram, expected) {
			t.Fatalf("telegram message missing %q: %q", expected, formatted.Telegram)
		}
	}
}
