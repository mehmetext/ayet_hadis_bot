package telegram

import "github.com/go-telegram/bot/models"

const telegramMessageLimit = 4096

func commandDefinitions() []models.BotCommand {
	return []models.BotCommand{
		{Command: "start", Description: "Abone ol ve bildirim dilini seç"},
		{Command: "stop", Description: "Bildirimleri durdur"},
		{Command: "language", Description: "Bildirim dilini değiştir"},
		{Command: "status", Description: "Abonelik durumunu gör"},
		{Command: "sample", Description: "Rastgele ayet veya hadis al"},
		{Command: "help", Description: "Kullanılabilir komutları göster"},
	}
}

func allowedUpdates() []string {
	return []string{"message", "callback_query"}
}
