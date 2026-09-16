package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/example/ayet-hadis-bot/internal/catalog"
)

type Service struct {
	Quran  QuranSynchronizer
	Hadith HadithSynchronizer
	Logger *log.Logger
}

func (service Service) SyncAll(ctx context.Context) {
	for _, language := range catalog.SupportedLanguages() {
		quranComplete, err := service.Quran.Store.IsSyncComplete(ctx, "quranenc", language.Code, language.QuranTranslationKey)
		if err != nil {
			service.Logger.Printf("sync warning: %v", err)
		} else if !quranComplete {
			if err := service.Quran.SyncLanguage(ctx, language); err != nil {
				service.Logger.Printf("sync warning: %v", err)
			}
		}
		hadithComplete, err := service.Hadith.Store.IsSyncComplete(ctx, "hadeethenc", language.Code, language.HadithLanguageKey)
		if err != nil {
			service.Logger.Printf("sync warning: %v", err)
		} else if !hadithComplete {
			if err := service.Hadith.SyncLanguage(ctx, language); err != nil {
				service.Logger.Printf("sync warning: %v", err)
			}
		}
	}
}
func (service Service) SyncOne(ctx context.Context, languageCode string) error {
	language, found := catalog.Find(languageCode)
	if !found {
		return fmt.Errorf("unsupported language: %s", languageCode)
	}
	if err := service.Quran.SyncLanguage(ctx, language); err != nil {
		return err
	}
	return service.Hadith.SyncLanguage(ctx, language)
}
