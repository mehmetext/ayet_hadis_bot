package catalog

import "testing"

func TestSupportedLanguagesHaveUniqueSourceKeys(t *testing.T) {
	languages := SupportedLanguages()
	if len(languages) != 4 {
		t.Fatalf("supported language count = %d, want 4", len(languages))
	}

	seenCodes := make(map[string]struct{}, len(languages))
	for _, language := range languages {
		if language.Code == "" || language.Name == "" {
			t.Fatalf("language metadata must be complete: %#v", language)
		}
		if language.QuranTranslationKey == "" || language.HadithLanguageKey == "" {
			t.Fatalf("source keys must be complete for %s", language.Code)
		}
		if _, exists := seenCodes[language.Code]; exists {
			t.Fatalf("duplicate language code %q", language.Code)
		}
		seenCodes[language.Code] = struct{}{}
	}
}
