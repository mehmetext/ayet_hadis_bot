package content

import (
	"testing"
)

func TestDecodeQuranResponse(t *testing.T) {
	input := []byte(`{"data":{"number":1,"text":"بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ","surah":{"number":1,"englishName":"Al-Fatihah","name":"سُورَةُ ٱلْفَاتِحَةِ"},"numberInSurah":1}}`)

	verse, err := DecodeQuranResponse(input)
	if err != nil {
		t.Fatalf("DecodeQuranResponse() error = %v", err)
	}
	if verse.Reference != "Al-Fatihah 1:1" {
		t.Fatalf("reference = %q, want %q", verse.Reference, "Al-Fatihah 1:1")
	}
	if verse.Text == "" {
		t.Fatal("expected verse text")
	}
}

func TestDecodeHadithResponse(t *testing.T) {
	input := []byte(`{"hadiths":[{"hadithnumber":1,"text":"Actions are judged by intentions.","reference":{"book":1,"hadith":1},"grades":[{"name":"Sahih","grade":"Sahih"}]}]}`)

	hadith, err := DecodeHadithResponse(input)
	if err != nil {
		t.Fatalf("DecodeHadithResponse() error = %v", err)
	}
	if hadith.Reference != "Book 1, Hadith 1" {
		t.Fatalf("reference = %q", hadith.Reference)
	}
	if hadith.Text == "" {
		t.Fatal("expected hadith text")
	}
}
