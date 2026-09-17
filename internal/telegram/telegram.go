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

type WelcomeSettings struct {
	Start      string
	End        string
	DailyCount int
}

type Bot struct {
	client  *bot.Bot
	store   UserStore
	logger  *log.Logger
	welcome WelcomeSettings
}

func New(token string, userStore UserStore, logger *log.Logger, welcome WelcomeSettings) (*Bot, error) {
	instance := &Bot{store: userStore, logger: logger, welcome: welcome}
	client, err := bot.New(token,
		bot.WithAllowedUpdates(allowedUpdates()),
		bot.WithErrorsHandler(func(err error) {
			logger.Printf("Telegram polling: %v", err)
		}),
		bot.WithMessageTextHandler("/start", bot.MatchTypePrefix, instance.start),
		bot.WithMessageTextHandler("/stop", bot.MatchTypeExact, instance.stop),
		bot.WithMessageTextHandler("/language", bot.MatchTypeExact, instance.language),
		bot.WithMessageTextHandler("/status", bot.MatchTypeExact, instance.status),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, instance.help),
		bot.WithCallbackQueryDataHandler("language:", bot.MatchTypePrefix, instance.selectLanguage),
	)
	if err != nil {
		return nil, fmt.Errorf("create Telegram bot: %w", err)
	}
	instance.client = client
	return instance, nil
}

func (instance *Bot) Initialize(ctx context.Context) error {
	if _, err := instance.client.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commandDefinitions()}); err != nil {
		return fmt.Errorf("set Telegram commands: %w", err)
	}
	return nil
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
		chunks := splitMessage(message, telegramMessageLimit)
		if err := instance.sendChunks(ctx, userID, chunks); err != nil {
			if isUserUnavailableError(err) {
				if unsubscribeErr := instance.store.UnsubscribeTelegramUser(ctx, userID); unsubscribeErr != nil {
					instance.logger.Printf("deactivate unavailable user=%d: %v", userID, unsubscribeErr)
				}
			}
			instance.logger.Printf("Telegram delivery user=%d language=%s: %v", userID, languageCode, err)
			continue
		}
		instance.logger.Printf("notification=sent user=%d language=%s parts=%d", userID, languageCode, len(chunks))
	}
	instance.logger.Printf("notification=finished language=%s recipients=%d", languageCode, len(users))
	return nil
}

func isUserUnavailableError(err error) bool {
	return errors.Is(err, bot.ErrorForbidden)
}

func (instance *Bot) sendChunks(ctx context.Context, userID int64, chunks []string) error {
	for _, chunk := range chunks {
		if err := instance.sendMessage(ctx, userID, chunk); err != nil {
			return err
		}
	}
	return nil
}

func splitMessage(message string, limit int) []string {
	runes := []rune(message)
	var chunks []string
	for len(runes) > limit {
		cut := limit
		for index := limit; index > 0; index-- {
			if runes[index-1] == '\n' || runes[index-1] == ' ' {
				cut = index
				break
			}
		}
		chunks = append(chunks, string(runes[:cut]))
		runes = runes[cut:]
	}
	if len(runes) > 0 {
		chunks = append(chunks, string(runes))
	}
	return chunks
}

func (instance *Bot) sendMessage(ctx context.Context, userID int64, message string) error {
	params := &bot.SendMessageParams{ChatID: userID, Text: message, ParseMode: models.ParseModeHTML}
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
	if update.Message == nil || update.Message.From == nil || !isStartCommand(update.Message.Text) {
		return
	}
	instance.logger.Printf("command=/start user=%d chat=%d", update.Message.From.ID, update.Message.Chat.ID)
	instance.sendLanguagePicker(ctx, client, update.Message.Chat.ID, welcomeMessage(instance.welcome))
}

func isStartCommand(text string) bool {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	command := strings.SplitN(fields[0], "@", 2)[0]
	return command == "/start"
}

func welcomeMessage(settings WelcomeSettings) string {
	return fmt.Sprintf("Ayet & Hadis Botu'na hoş geldin.\n\nBu bot, QuranEnc ve HadeethEnc kaynaklarından ayet ve hadisleri anlık olarak paylaşır.\n\nHer gün %s–%s arasında %d bildirim göndeririz. Bildirimler ayet ve hadis sırayla olacak şekilde ilerler.\n\nKomutlar:\n/start — Abone ol ve dil seç\n/stop — Bildirimleri durdur\n/language — Bildirim dilini değiştir\n/status — Abonelik durumunu gör\n/help — Bu yardım mesajını göster\n\nBaşlamak için bildirim dilini seç:", settings.Start, settings.End, settings.DailyCount)
}

func (instance *Bot) help(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	instance.logger.Printf("command=/help user=%d chat=%d", update.Message.From.ID, update.Message.Chat.ID)
	if _, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: helpMessage()}); err != nil {
		instance.logger.Printf("Telegram help response: %v", err)
	}
}

func helpMessage() string {
	return "Komutlar:\n/start — Abone ol ve dil seç\n/stop — Bildirimleri durdur\n/language — Bildirim dilini değiştir\n/status — Abonelik durumunu gör\n/help — Bu yardım mesajını göster"
}

func (instance *Bot) language(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	instance.logger.Printf("command=/language user=%d chat=%d", update.Message.From.ID, update.Message.Chat.ID)
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
	if _, err := client.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
		instance.logger.Printf("answer language callback: %v", err)
	}
	languageCode := strings.TrimPrefix(query.Data, "language:")
	if _, found := catalog.Find(languageCode); !found {
		return
	}
	if err := instance.store.SaveTelegramUser(ctx, query.From.ID, languageCode); err != nil {
		instance.logger.Printf("save Telegram user: %v", err)
		return
	}
	instance.logger.Printf("subscription=active user=%d language=%s", query.From.ID, languageCode)
	_, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: query.From.ID, Text: fmt.Sprintf("%s dilinde bildirimlere abone oldun.", languageName(languageCode))})
	if err != nil {
		instance.logger.Printf("Telegram language confirmation: %v", err)
	}
}

func (instance *Bot) stop(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}
	instance.logger.Printf("command=/stop user=%d chat=%d", update.Message.From.ID, update.Message.Chat.ID)
	if err := instance.store.UnsubscribeTelegramUser(ctx, update.Message.From.ID); err != nil {
		instance.logger.Printf("unsubscribe Telegram user: %v", err)
		return
	}
	instance.logger.Printf("subscription=inactive user=%d", update.Message.From.ID)
	_, err := client.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: "Bildirim aboneliğin durduruldu. Yeniden başlamak için /start yazabilirsin."})
	if err != nil {
		instance.logger.Printf("Telegram stop confirmation: %v", err)
	}
}

func (instance *Bot) status(ctx context.Context, client *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}
	instance.logger.Printf("command=/status user=%d chat=%d", update.Message.From.ID, update.Message.Chat.ID)
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
