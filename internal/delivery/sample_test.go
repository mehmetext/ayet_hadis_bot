package delivery

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/mehmetext/ayet-hadis-bot/internal/selection"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/quranenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/store"
)

type sampleQuranClient struct{}

func (sampleQuranClient) FetchAyah(context.Context, string, int, int) (quranenc.Verse, error) {
	return quranenc.Verse{SurahNumber: 1, AyahNumber: 1, ArabicText: "Ayet metni", Translation: "Verse text"}, nil
}

type sampleHadithSelector struct {
	Random *rand.Rand
}

func (sampleHadithSelector) QuranCandidate(string) (store.Candidate, error) {
	return store.Candidate{}, nil
}
func (sampleHadithSelector) HadithCandidate(context.Context, string, string) (store.Candidate, error) {
	return store.Candidate{LanguageCode: "tur", Type: store.ContentTypeHadith, Source: "test", ID: "hadith-1"}, nil
}
func (sampleHadithSelector) AllHadithCandidates(context.Context, string, string) ([]store.Candidate, error) {
	return nil, nil
}

type sampleHadithClient struct{}

func (sampleHadithClient) One(context.Context, string, string) (hadeethenc.Hadith, error) {
	return hadeethenc.Hadith{Reference: "Buhari 1", Hadeeth: "Hadis metni", Grade: "Sahih"}, nil
}

func TestSampleReturnsContentWithoutChangingDeliveryState(t *testing.T) {
	database, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	service := Service{
		Store:    database,
		Selector: selection.Selector{Random: rand.New(rand.NewSource(0))},
		Random:   rand.New(rand.NewSource(0)),
		Quran:    sampleQuranClient{},
	}

	message, err := service.Sample(context.Background(), "tur")
	if err != nil {
		t.Fatalf("Sample returned error: %v", err)
	}
	if !strings.Contains(message, "AYET") {
		t.Fatalf("sample message does not contain a content heading: %q", message)
	}
	nextType, err := database.NextType(context.Background(), "tur")
	if err != nil {
		t.Fatal(err)
	}
	if nextType != store.ContentTypeVerse {
		t.Fatalf("next content type = %q, want verse", nextType)
	}
	for _, contentType := range []store.ContentType{store.ContentTypeVerse, store.ContentTypeHadith} {
		count, err := database.HistoryCount(context.Background(), "tur", contentType)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("history count for %s = %d, want 0", contentType, count)
		}
	}
}

func TestSampleFetchesHadithWithoutChangingDeliveryState(t *testing.T) {
	database, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	service := Service{Store: database, Selector: sampleHadithSelector{Random: rand.New(rand.NewSource(1))}, Random: rand.New(rand.NewSource(1)), Hadith: sampleHadithClient{}}

	message, err := service.Sample(context.Background(), "tur")
	if err != nil {
		t.Fatalf("Sample returned error: %v", err)
	}
	if !strings.Contains(message, "HADİS") || !strings.Contains(message, "Buhari 1") {
		t.Fatalf("sample message does not contain hadith content: %q", message)
	}
	count, err := database.HistoryCount(context.Background(), "tur", store.ContentTypeHadith)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("hadith history count = %d, want 0", count)
	}
}
