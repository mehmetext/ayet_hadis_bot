package store

import (
	"context"
	"testing"
	"time"
)

func TestCompletionAdvancesTypeOnlyAfterSuccess(t *testing.T) {
	database := openTestDatabase(t)
	ctx := context.Background()
	candidate := Candidate{LanguageCode: "tur", Type: ContentTypeVerse, Source: "quranenc", ID: "2:255"}
	if next, _ := database.NextType(ctx, "tur"); next != ContentTypeVerse {
		t.Fatalf("next = %s", next)
	}
	reserved, err := database.Reserve(ctx, candidate, time.Now().Add(time.Minute))
	if err != nil || !reserved {
		t.Fatalf("Reserve() = %t, %v", reserved, err)
	}
	if next, _ := database.NextType(ctx, "tur"); next != ContentTypeVerse {
		t.Fatalf("reservation advanced type: %s", next)
	}
	if err := database.Complete(ctx, candidate, time.Now()); err != nil {
		t.Fatal(err)
	}
	if next, _ := database.NextType(ctx, "tur"); next != ContentTypeHadith {
		t.Fatalf("next = %s", next)
	}
}

func TestHistoryBlocksDuplicateReservation(t *testing.T) {
	database := openTestDatabase(t)
	ctx := context.Background()
	candidate := Candidate{LanguageCode: "tur", Type: ContentTypeVerse, Source: "quranenc", ID: "1:1"}
	reserved, _ := database.Reserve(ctx, candidate, time.Now().Add(time.Minute))
	if !reserved {
		t.Fatal("initial reservation failed")
	}
	if err := database.Complete(ctx, candidate, time.Now()); err != nil {
		t.Fatal(err)
	}
	reserved, err := database.Reserve(ctx, candidate, time.Now().Add(time.Minute))
	if err != nil || reserved {
		t.Fatalf("duplicate reserve = %t, %v", reserved, err)
	}
}

func TestSchemaDoesNotCreateContentArchiveTables(t *testing.T) {
	database := openTestDatabase(t)
	for _, table := range []string{"quran_verses", "hadiths", "quran_editions", "hadith_editions"} {
		var count int
		if err := database.database.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("archive table %q exists", table)
		}
	}
}

func openTestDatabase(t *testing.T) *Store {
	t.Helper()
	database, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}
