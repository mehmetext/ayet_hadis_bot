package telegram

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func sampleLanguage(ctx context.Context, userStore UserStore, userID int64) (string, error) {
	user, found, err := userStore.TelegramUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if !found || user.LanguageCode == "" {
		return "", errTelegramUserNotRegistered
	}
	return user.LanguageCode, nil
}

func (instance *Bot) sample(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil || instance.sampler == nil {
		return
	}
	userID := update.Message.From.ID
	instance.logger.Printf("command=/sample user=%d chat=%d", userID, update.Message.Chat.ID)
	languageCode, err := sampleLanguage(ctx, instance.store, userID)
	if errors.Is(err, errTelegramUserNotRegistered) {
		instance.sendSampleMessage(ctx, client, update.Message.Chat.ID, "Önce bildirim dilini seçmek için /start komutunu kullanmalısın.")
		return
	}
	if err != nil {
		instance.logger.Printf("read Telegram user for sample: %v", err)
		instance.sendSampleMessage(ctx, client, update.Message.Chat.ID, "Örnek içerik alınırken bir hata oluştu. Lütfen biraz sonra tekrar dene.")
		return
	}
	message, err := instance.sampler.Sample(ctx, languageCode)
	if err != nil {
		instance.logger.Printf("sample content language=%s: %v", languageCode, err)
		instance.sendSampleMessage(ctx, client, update.Message.Chat.ID, "Örnek içerik alınırken bir hata oluştu. Lütfen biraz sonra tekrar dene.")
		return
	}
	if err := instance.sendChunks(ctx, update.Message.Chat.ID, splitMessage(message, telegramMessageLimit)); err != nil {
		instance.logger.Printf("Telegram sample response: %v", err)
	}
}

func (instance *Bot) sendSampleMessage(ctx context.Context, client *bot.Bot, chatID int64, message string) {
	if _, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: message}); err != nil {
		instance.logger.Printf("Telegram sample response: %v", err)
	}
}
