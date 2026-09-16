package store

import (
	"context"
	"testing"
)

func TestNextContentTypeAlternatesPersistently(t *testing.T) {
	database := openTestDatabase(t)
	ctx := context.Background()

	first, err := database.NextContentType(ctx)
	if err != nil {
		t.Fatalf("NextContentType() error = %v", err)
	}
	if first != ContentTypeVerse {
		t.Fatalf("first type = %q, want %q", first, ContentTypeVerse)
	}
	second, err := database.NextContentType(ctx)
	if err != nil {
		t.Fatalf("NextContentType() error = %v", err)
	}
	if second != ContentTypeHadith {
		t.Fatalf("second type = %q, want %q", second, ContentTypeHadith)
	}
}

func openTestDatabase(t *testing.T) *Store {
	t.Helper()
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}
