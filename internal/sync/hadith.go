package sync

import (
	"context"
	"fmt"

	"github.com/example/ayet-hadis-bot/internal/catalog"
	"github.com/example/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/example/ayet-hadis-bot/internal/store"
)

const hadithPageSize = 100

type HadithSynchronizer struct {
	Client hadeethenc.Client
	Store  *store.Store
}

func (s HadithSynchronizer) SyncLanguage(ctx context.Context, language catalog.Language) error {
	categories, err := s.Client.RootCategories(ctx, language.HadithLanguageKey)
	if err != nil {
		return err
	}
	ids := make(map[string]struct{})
	for _, category := range categories {
		for page := 1; ; page++ {
			listings, lastPage, err := s.Client.List(ctx, language.HadithLanguageKey, category.ID, page, hadithPageSize)
			if err != nil {
				return fmt.Errorf("list hadith category %s page %d: %w", category.ID, page, err)
			}
			for _, listing := range listings {
				ids[listing.ID] = struct{}{}
			}
			if page >= lastPage {
				break
			}
		}
	}
	hadiths := make([]store.Hadith, 0, len(ids))
	for id := range ids {
		item, err := s.Client.One(ctx, language.HadithLanguageKey, id)
		if err != nil {
			return fmt.Errorf("fetch hadith %s: %w", id, err)
		}
		hadiths = append(hadiths, store.Hadith{ID: item.ID, CollectionName: "HadeethEnc", Reference: item.Attribution, Text: item.Hadeeth, Grade: item.Grade, Explanation: item.Explanation})
	}
	return s.Store.ReplaceHadithEdition(ctx, language.Code, language.HadithLanguageKey, language.Publisher, language.Attribution, "HadeethEnc API v1", hadiths)
}
