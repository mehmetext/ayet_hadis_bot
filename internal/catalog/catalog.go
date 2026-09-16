package catalog

type Language struct {
	Code                string
	QuranTranslationKey string
	HadithLanguageKey   string
	Attribution         string
}

var supportedLanguages = []Language{
	{
		Code: "ara",
		// QuranEnc returns the canonical Arabic text alongside each translation.
		// Arabic delivery uses that preserved field rather than a translated edition.
		QuranTranslationKey: "english_rwwad",
		HadithLanguageKey:   "ar",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "eng",
		QuranTranslationKey: "english_rwwad",
		HadithLanguageKey:   "en",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "tur",
		QuranTranslationKey: "turkish_rwwad",
		HadithLanguageKey:   "tr",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "deu",
		QuranTranslationKey: "german_rwwad",
		HadithLanguageKey:   "de",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
}

func Find(code string) (Language, bool) {
	for _, language := range supportedLanguages {
		if language.Code == code {
			return language, true
		}
	}
	return Language{}, false
}
