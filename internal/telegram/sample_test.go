package telegram

import (
	"context"
	"errors"
	"testing"

	"github.com/mehmetext/ayet-hadis-bot/internal/store"
)

type sampleUserStore struct {
	user  store.TelegramUser
	found bool
	err   error
}

func (sampleUserStore) SaveTelegramUser(context.Context, int64, string) error { return nil }
func (sampleUserStore) UnsubscribeTelegramUser(context.Context, int64) error  { return nil }
func (sampleUserStore) SubscribedTelegramUsers(context.Context, string) ([]int64, error) {
	return nil, nil
}
func (store sampleUserStore) TelegramUser(context.Context, int64) (store.TelegramUser, bool, error) {
	return store.user, store.found, store.err
}

func TestSampleLanguageReturnsRegisteredUserLanguage(t *testing.T) {
	userStore := sampleUserStore{user: store.TelegramUser{LanguageCode: "tur"}, found: true}

	language, err := sampleLanguage(context.Background(), userStore, 42)

	if err != nil {
		t.Fatalf("sampleLanguage returned error: %v", err)
	}
	if language != "tur" {
		t.Fatalf("language = %q, want tur", language)
	}
}

func TestSampleLanguageRejectsUnregisteredUser(t *testing.T) {
	userStore := sampleUserStore{found: false}

	_, err := sampleLanguage(context.Background(), userStore, 42)

	if !errors.Is(err, errTelegramUserNotRegistered) {
		t.Fatalf("error = %v, want %v", err, errTelegramUserNotRegistered)
	}
}
