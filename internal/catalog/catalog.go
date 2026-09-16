package catalog

type Direction string

const (
	DirectionLeftToRight Direction = "ltr"
	DirectionRightToLeft Direction = "rtl"
)

type Language struct {
	Code                string
	Name                string
	QuranTranslationKey string
	HadithLanguageKey   string
	Direction           Direction
	Publisher           string
	Attribution         string
}

var supportedLanguages = []Language{
	{
		Code: "ara",
		Name: "Arabic",
		// QuranEnc returns the canonical Arabic text alongside each translation.
		// Arabic delivery uses that preserved field rather than a translated edition.
		QuranTranslationKey: "english_rwwad",
		HadithLanguageKey:   "ar",
		Direction:           DirectionRightToLeft,
		Publisher:           "QuranEnc and HadeethEnc",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "eng",
		Name:                "English",
		QuranTranslationKey: "english_rwwad",
		HadithLanguageKey:   "en",
		Direction:           DirectionLeftToRight,
		Publisher:           "QuranEnc and HadeethEnc",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "tur",
		Name:                "Turkish",
		QuranTranslationKey: "turkish_rwwad",
		HadithLanguageKey:   "tr",
		Direction:           DirectionLeftToRight,
		Publisher:           "QuranEnc and HadeethEnc",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
	{
		Code:                "deu",
		Name:                "German",
		QuranTranslationKey: "german_rwwad",
		HadithLanguageKey:   "de",
		Direction:           DirectionLeftToRight,
		Publisher:           "QuranEnc and HadeethEnc",
		Attribution:         "QuranEnc.com; HadeethEnc.com",
	},
}

func SupportedLanguages() []Language {
	languages := make([]Language, len(supportedLanguages))
	copy(languages, supportedLanguages)
	return languages
}

func Find(code string) (Language, bool) {
	for _, language := range supportedLanguages {
		if language.Code == code {
			return language, true
		}
	}
	return Language{}, false
}
