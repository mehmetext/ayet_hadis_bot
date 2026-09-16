package selection

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/mehmetext/ayet-hadis-bot/internal/catalog"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/store"
)

const (
	quranSource    = "quranenc"
	hadithSource   = "hadeethenc"
	hadithPageSize = 50
)

type Selector struct {
	Random *rand.Rand
	Hadith hadeethenc.Client
}

func (selector Selector) QuranCandidate(languageCode string) (store.Candidate, error) {
	surah, ayah, ok := catalog.VerseCoordinate(selector.Random.Intn(catalog.QuranVerseCount))
	if !ok {
		return store.Candidate{}, fmt.Errorf("generate Quran coordinate")
	}
	return store.Candidate{LanguageCode: languageCode, Type: store.ContentTypeVerse, Source: quranSource, ID: fmt.Sprintf("%d:%d", surah, ayah)}, nil
}
func (selector Selector) HadithCandidate(ctx context.Context, languageCode, hadithLanguage string) (store.Candidate, error) {
	categories, err := selector.Hadith.RootCategories(ctx, hadithLanguage)
	if err != nil {
		return store.Candidate{}, err
	}
	if len(categories) == 0 {
		return store.Candidate{}, fmt.Errorf("no HadeethEnc categories for %s", hadithLanguage)
	}
	category := categories[selector.Random.Intn(len(categories))]
	listings, lastPage, err := selector.Hadith.List(ctx, hadithLanguage, category.ID, 1, hadithPageSize)
	if err != nil {
		return store.Candidate{}, err
	}
	if lastPage > 1 {
		page := selector.Random.Intn(lastPage) + 1
		listings, _, err = selector.Hadith.List(ctx, hadithLanguage, category.ID, page, hadithPageSize)
		if err != nil {
			return store.Candidate{}, err
		}
	}
	if len(listings) == 0 {
		return store.Candidate{}, fmt.Errorf("empty HadeethEnc category page")
	}
	return store.Candidate{LanguageCode: languageCode, Type: store.ContentTypeHadith, Source: hadithSource, ID: listings[selector.Random.Intn(len(listings))].ID}, nil
}

func (selector Selector) AllHadithCandidates(ctx context.Context, languageCode, hadithLanguage string) ([]store.Candidate, error) {
	categories, err := selector.Hadith.RootCategories(ctx, hadithLanguage)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, category := range categories {
		for page := 1; ; page++ {
			listings, lastPage, err := selector.Hadith.List(ctx, hadithLanguage, category.ID, page, hadithPageSize)
			if err != nil {
				return nil, err
			}
			for _, listing := range listings {
				seen[listing.ID] = struct{}{}
			}
			if page >= lastPage {
				break
			}
		}
	}
	candidates := make([]store.Candidate, 0, len(seen))
	for id := range seen {
		candidates = append(candidates, store.Candidate{LanguageCode: languageCode, Type: store.ContentTypeHadith, Source: hadithSource, ID: id})
	}
	return candidates, nil
}
