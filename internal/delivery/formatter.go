package delivery

import (
	"fmt"
	"html"

	"github.com/mehmetext/ayet-hadis-bot/internal/catalog"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/quranenc"
)

type formattedContent struct {
	Telegram string
	Console  string
}

func formatVerse(language catalog.Language, verse quranenc.Verse) formattedContent {
	surahName, _ := catalog.SurahName(verse.SurahNumber)
	if language.Code == "ara" {
		return formattedContent{
			Console:  fmt.Sprintf("📖 AYET\n\n%s Suresi — %d:%d\n\n%s\n\nKaynak: %s\n", surahName, verse.SurahNumber, verse.AyahNumber, verse.ArabicText, language.Attribution),
			Telegram: fmt.Sprintf("<b>📖 AYET</b>\n\n<b>%s Suresi — %d:%d</b>\n\n%s\n\nKaynak: %s\n", html.EscapeString(surahName), verse.SurahNumber, verse.AyahNumber, html.EscapeString(verse.ArabicText), html.EscapeString(language.Attribution)),
		}
	}
	return formattedContent{
		Console:  fmt.Sprintf("📖 AYET\n\n%s Suresi — %d:%d\n\n%s\n\nDil çevirisi:\n%s\n\nKaynak: %s\n", surahName, verse.SurahNumber, verse.AyahNumber, verse.ArabicText, verse.Translation, language.Attribution),
		Telegram: fmt.Sprintf("<b>📖 AYET</b>\n\n<b>%s Suresi — %d:%d</b>\n\n%s\n\n<b>Dil çevirisi:</b>\n%s\n\nKaynak: %s\n", html.EscapeString(surahName), verse.SurahNumber, verse.AyahNumber, html.EscapeString(verse.ArabicText), html.EscapeString(verse.Translation), html.EscapeString(language.Attribution)),
	}
}

func formatHadith(language catalog.Language, hadith hadeethenc.Hadith) formattedContent {
	reference := hadith.Reference
	if reference == "" {
		reference = hadith.Attribution
	}
	return formattedContent{
		Console:  fmt.Sprintf("📜 HADİS\n\nReferans: %s\n\n%s\n\nDerece: %s\n\nAçıklama:\n%s\n\nKaynak: %s\n", reference, hadith.Hadeeth, hadith.Grade, hadith.Explanation, language.Attribution),
		Telegram: fmt.Sprintf("<b>📜 HADİS</b>\n\n<b>Referans:</b> %s\n\n%s\n\n<b>Derece:</b> %s\n\n<b>Açıklama:</b>\n%s\n\nKaynak: %s\n", html.EscapeString(reference), html.EscapeString(hadith.Hadeeth), html.EscapeString(hadith.Grade), html.EscapeString(hadith.Explanation), html.EscapeString(language.Attribution)),
	}
}
