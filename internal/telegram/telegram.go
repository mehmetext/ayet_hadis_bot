package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mehmetext/ayet-hadis-bot/internal/catalog"
	"github.com/mehmetext/ayet-hadis-bot/internal/store"
)

type UserStore interface {
	SaveTelegramUser(context.Context, int64, string) error
	UnsubscribeTelegramUser(context.Context, int64) error
	TelegramUser(context.Context, int64) (store.TelegramUser, bool, error)
	SubscribedTelegramUsers(context.Context, string) ([]int64, error)
}

type Bot struct {
	client *bot.Bot
	store  UserStore
	logger *log.Logger
}

func New(token string, userStore UserStore, logger *log.Logger) (*Bot, error) {
	instance := &Bot{store: userStore, logger: logger}
	client, err := bot.New(token,
		bot.WithMessageTextHandler("/start", bot.MatchTypePrefix, instance.start),
		bot.WithMessageTextHandler("/stop", bot.MatchTypeExact, instance.stop),
		bot.WithMessageTextHandler("/language", bot.MatchTypeExact, instance.language),
		bot.WithMessageTextHandler("/status", bot.MatchTypeExact, instance.status),
		bot.WithCallbackQueryDataHandler("language:", bot.MatchTypePrefix, instance.selectLanguage),
	)
	if err != nil {
		return nil, fmt.Errorf("create Telegram bot: %w", err)
	}
	instance.client = client
	return instance, nil
}

func (instance *Bot) Start(ctx context.Context) { instance.client.Start(ctx) }

// Broadcast sends one language's shared content to every active subscriber.
// Individual Telegram failures are logged so one blocked user does not stop the language stream.
func (instance *Bot) Broadcast(ctx context.Context, languageCode, message string) error {
	users, err := instance.store.SubscribedTelegramUsers(ctx, languageCode)
	if err != nil {
		return err
	}
	for _, userID := range users {
		if err := instance.sendMessage(ctx, userID, message); err != nil {
			instance.logger.Printf("Telegram delivery user=%d language=%s: %v", userID, languageCode, err)
		}
	}
	return nil
}

func (instance *Bot) sendMessage(ctx context.Context, userID int64, message string) error {
	params := &bot.SendMessageParams{ChatID: userID, Text: message}
	if _, err := instance.client.SendMessage(ctx, params); err == nil {
		return nil
	} else {
		var rateLimit *bot.TooManyRequestsError
		if !errors.As(err, &rateLimit) || rateLimit.RetryAfter <= 0 {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(rateLimit.RetryAfter) * time.Second):
		}
	}
	_, err := instance.client.SendMessage(ctx, params)
	return err
}

func (instance *Bot) start(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	instance.sendLanguagePicker(ctx, client, update.Message.Chat.ID, "Dilini seç; seçimden sonra bildirimler başlayacak.")
}

func (instance *Bot) language(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	instance.sendLanguagePicker(ctx, client, update.Message.Chat.ID, "Bildirim dilini seç:")
}

func (instance *Bot) sendLanguagePicker(ctx context.Context, client *bot.Bot, chatID int64, text string) {
	_, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ReplyMarkup: languageKeyboard()})
	if err != nil {
		instance.logger.Printf("Telegram language picker: %v", err)
	}
}

func (instance *Bot) selectLanguage(ctx context.Context, client *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}
	_, _ = client.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})
	languageCode := strings.TrimPrefix(query.Data, "language:")
	if _, found := catalog.Find(languageCode); !found {
		return
	}
	if err := instance.store.SaveTelegramUser(ctx, query.From.ID, languageCode); err != nil {
		instance.logger.Printf("save Telegram user: %v", err)
		return
	}
	_, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: query.From.ID, Text: fmt.Sprintf("%s dilinde bildirimlere abone oldun.", languageName(languageCode))})
	if err != nil {
		instance.logger.Printf("Telegram language confirmation: %v", err)
	}
}

func (instance *Bot) stop(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}
	if err := instance.store.UnsubscribeTelegramUser(ctx, update.Message.From.ID); err != nil {
		instance.logger.Printf("unsubscribe Telegram user: %v", err)
		return
	}
	_, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: "Bildirim aboneliğin durduruldu. Yeniden başlamak için /start yazabilirsin."})
	if err != nil {
		instance.logger.Printf("Telegram stop confirmation: %v", err)
	}
}

func (instance *Bot) status(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}
	user, found, err := instance.store.TelegramUser(ctx, update.Message.From.ID)
	if err != nil {
		instance.logger.Printf("read Telegram user: %v", err)
		return
	}
	text := "Henüz kayıtlı değilsin. Başlamak için /start yaz."
	if found {
		state := "pasif"
		if user.Subscribed {
			state = "aktif"
		}
		text = fmt.Sprintf("Dil: %s\nAbonelik: %s", languageName(user.LanguageCode), state)
	}
	if _, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: text}); err != nil {
		instance.logger.Printf("Telegram status response: %v", err)
	}
}

func languageKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{
		{Text: "العربية", CallbackData: "language:ara"},
		{Text: "English", CallbackData: "language:eng"},
	}, {
		{Text: "Türkçe", CallbackData: "language:tur"},
		{Text: "Deutsch", CallbackData: "language:deu"},
	}}}
}

func languageName(code string) string {
	switch code {
	case "ara":
		return "Arapça"
	case "eng":
		return "İngilizce"
	case "tur":
		return "Türkçe"
	case "deu":
		return "Almanca"
	default:
		return code
	}
}
