package delivery

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/ayet-hadis-bot/internal/catalog"
	"github.com/example/ayet-hadis-bot/internal/selection"
	"github.com/example/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/example/ayet-hadis-bot/internal/source/quranenc"
	"github.com/example/ayet-hadis-bot/internal/store"
)

const reservationDuration = 24 * time.Hour
const candidateAttempts = 100

type Service struct {
	Store       *store.Store
	Selector    selection.Selector
	Quran       quranenc.Client
	Hadith      hadeethenc.Client
	Writer      io.Writer
	Broadcaster func(context.Context, string, string) error
}

func (service Service) Attempt(ctx context.Context, languageCode string, slot, now time.Time) error {
	language, found := catalog.Find(languageCode)
	if !found {
		return fmt.Errorf("unsupported language: %s", languageCode)
	}
	pending, exists, err := service.Store.Pending(ctx, languageCode)
	if err != nil {
		return err
	}
	if exists && pending.NextAttemptAt.After(now) {
		return nil
	}
	if exists {
		return service.deliver(ctx, language, pending.Candidate, pending.Slot, pending.Attempts, now)
	}
	contentType, err := service.Store.NextType(ctx, languageCode)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < candidateAttempts; attempt++ {
		candidate, err := service.candidate(ctx, language, contentType)
		if err != nil {
			return err
		}
		reserved, err := service.Store.Reserve(ctx, candidate, now.Add(reservationDuration))
		if err != nil {
			return err
		}
		if !reserved {
			continue
		}
		return service.deliver(ctx, language, candidate, slot, 0, now)
	}
	if contentType == store.ContentTypeVerse {
		count, err := service.Store.HistoryCount(ctx, languageCode, contentType)
		if err != nil {
			return err
		}
		if count >= catalog.QuranVerseCount {
			if err := service.Store.ClearCycle(ctx, languageCode, contentType); err != nil {
				return err
			}
			return service.Attempt(ctx, languageCode, slot, now)
		}
	}
	if contentType == store.ContentTypeHadith {
		candidates, err := service.Selector.AllHadithCandidates(ctx, languageCode, language.HadithLanguageKey)
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			reserved, err := service.Store.Reserve(ctx, candidate, now.Add(reservationDuration))
			if err != nil {
				return err
			}
			if reserved {
				return service.deliver(ctx, language, candidate, slot, 0, now)
			}
		}
		if err := service.Store.ClearCycle(ctx, languageCode, contentType); err != nil {
			return err
		}
		return service.Attempt(ctx, languageCode, slot, now)
	}
	return fmt.Errorf("no unreserved %s candidate after %d attempts", contentType, candidateAttempts)
}

func (service Service) candidate(ctx context.Context, language catalog.Language, contentType store.ContentType) (store.Candidate, error) {
	if contentType == store.ContentTypeVerse {
		return service.Selector.QuranCandidate(language.Code)
	}
	return service.Selector.HadithCandidate(ctx, language.Code, language.HadithLanguageKey)
}
func (service Service) deliver(ctx context.Context, language catalog.Language, candidate store.Candidate, slot time.Time, attempts int, now time.Time) error {
	if candidate.Type == store.ContentTypeVerse {
		return service.deliverVerse(ctx, language, candidate, slot, attempts, now)
	}
	return service.deliverHadith(ctx, language, candidate, slot, attempts, now)
}
func (service Service) deliverVerse(ctx context.Context, language catalog.Language, candidate store.Candidate, slot time.Time, attempts int, now time.Time) error {
	var surah, ayah int
	if _, err := fmt.Sscanf(candidate.ID, "%d:%d", &surah, &ayah); err != nil {
		return err
	}
	verse, err := service.Quran.FetchAyah(ctx, language.QuranTranslationKey, surah, ayah)
	if err != nil {
		return service.failed(ctx, candidate, slot, attempts, now, err)
	}
	message := ""
	if language.Code == "ara" {
		message = fmt.Sprintf("AYET — %d:%d\n%s\nKaynak: %s\n", verse.SurahNumber, verse.AyahNumber, verse.ArabicText, language.Attribution)
	} else {
		message = fmt.Sprintf("AYET — %d:%d\n%s\n%s\nKaynak: %s\n", verse.SurahNumber, verse.AyahNumber, verse.ArabicText, verse.Translation, language.Attribution)
	}
	err = service.write(ctx, language.Code, message)
	if err != nil {
		return service.failed(ctx, candidate, slot, attempts, now, err)
	}
	return service.Store.Complete(ctx, candidate, slot)
}
func (service Service) deliverHadith(ctx context.Context, language catalog.Language, candidate store.Candidate, slot time.Time, attempts int, now time.Time) error {
	hadith, err := service.Hadith.One(ctx, language.HadithLanguageKey, candidate.ID)
	if err != nil {
		return service.failed(ctx, candidate, slot, attempts, now, err)
	}
	reference := hadith.Reference
	if reference == "" {
		reference = hadith.Attribution
	}
	message := fmt.Sprintf("HADİS — %s\n%s\nDerece: %s\nAçıklama: %s\nKaynak: %s\n", reference, hadith.Hadeeth, hadith.Grade, hadith.Explanation, language.Attribution)
	err = service.write(ctx, language.Code, message)
	if err != nil {
		return service.failed(ctx, candidate, slot, attempts, now, err)
	}
	return service.Store.Complete(ctx, candidate, slot)
}

func (service Service) write(ctx context.Context, languageCode, message string) error {
	if service.Broadcaster != nil {
		return service.Broadcaster(ctx, languageCode, message)
	}
	if service.Writer == nil {
		return fmt.Errorf("delivery output is not configured")
	}
	_, err := io.WriteString(service.Writer, message)
	return err
}
func (service Service) failed(ctx context.Context, candidate store.Candidate, slot time.Time, attempts int, now time.Time, cause error) error {
	if !retryable(cause) {
		_ = service.Store.Release(ctx, candidate)
		return cause
	}
	attempts++
	delay := retryDelay(attempts)
	if err := service.Store.SavePending(ctx, store.Pending{Candidate: candidate, Slot: slot, Attempts: attempts, NextAttemptAt: now.Add(delay)}); err != nil {
		return err
	}
	return fmt.Errorf("delivery deferred for %s: %w", delay, cause)
}
func retryDelay(attempt int) time.Duration {
	if attempt == 1 {
		return time.Minute
	}
	if attempt == 2 {
		return 5 * time.Minute
	}
	if attempt == 3 {
		return 15 * time.Minute
	}
	return 30 * time.Minute
}
func retryable(err error) bool {
	text := err.Error()
	return !strings.Contains(text, "status: 4") || strings.Contains(text, "status: 429")
}
