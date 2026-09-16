package sync

import (
	"context"
	"fmt"

	"github.com/example/ayet-hadis-bot/internal/catalog"
	"github.com/example/ayet-hadis-bot/internal/source/quranenc"
	"github.com/example/ayet-hadis-bot/internal/store"
)

const quranVerseCount = 6236

type QuranSynchronizer struct {
	Client quranenc.Client
	Store  *store.Store
}

func (s QuranSynchronizer) SyncLanguage(ctx context.Context, language catalog.Language) error {
	translations, err := s.Client.AvailableTranslations(ctx, quranLanguageCode(language.Code))
	if err != nil {
		return err
	}
	var version string
	available := false
	for _, translation := range translations {
		if translation.Key == language.QuranTranslationKey {
			available, version = true, translation.Version
			break
		}
	}
	if !available {
		return fmt.Errorf("%s Quran translation %q is unavailable", language.Name, language.QuranTranslationKey)
	}
	verses := make([]store.QuranVerse, 0, quranVerseCount)
	for surahNumber := 1; surahNumber <= 114; surahNumber++ {
		surah, err := s.Client.FetchSurah(ctx, language.QuranTranslationKey, surahNumber)
		if err != nil {
			return fmt.Errorf("fetch Quran surah %d: %w", surahNumber, err)
		}
		for _, verse := range surah {
			text := verse.Translation
			if language.Code == "ara" {
				text = verse.ArabicText
			}
			verses = append(verses, store.QuranVerse{SurahNumber: verse.SurahNumber, AyahNumber: verse.AyahNumber, ArabicText: verse.ArabicText, Text: text})
		}
	}
	return s.Store.ReplaceQuranEdition(ctx, language.Code, language.QuranTranslationKey, language.Publisher, language.Attribution, version, verses)
}
func quranLanguageCode(code string) string {
	if code == "ara" {
		return "en"
	}
	if code == "eng" {
		return "en"
	}
	if code == "tur" {
		return "tr"
	}
	return "de"
}
