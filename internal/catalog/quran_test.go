package catalog

import "testing"

func TestVerseCoordinateCoversQuran(t *testing.T) {
	surah, ayah, ok := VerseCoordinate(QuranVerseCount - 1)
	if !ok || surah != 114 || ayah != 6 {
		t.Fatalf("final coordinate = %d:%d, %t", surah, ayah, ok)
	}
	if _, _, ok := VerseCoordinate(QuranVerseCount); ok {
		t.Fatal("out-of-range coordinate is valid")
	}
}
