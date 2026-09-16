package content

type Verse struct {
	Number        int
	SurahName     string
	NumberInSurah int
	Text          string
	Reference     string
}

type Hadith struct {
	Number    int
	Text      string
	Reference string
	Grade     string
}
