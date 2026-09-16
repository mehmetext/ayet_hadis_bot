package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"math/rand"

	"github.com/mehmetext/ayet-hadis-bot/internal/config"
	"github.com/mehmetext/ayet-hadis-bot/internal/delivery"
	"github.com/mehmetext/ayet-hadis-bot/internal/schedule"
	"github.com/mehmetext/ayet-hadis-bot/internal/selection"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/source/quranenc"
	"github.com/mehmetext/ayet-hadis-bot/internal/store"
	"github.com/mehmetext/ayet-hadis-bot/internal/telegram"
)

func main() {
	logger := log.New(os.Stderr, "ayet-hadis-bot: ", log.LstdFlags)
	configuration, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}
	if err := os.MkdirAll(configuration.DataDirectory, 0o755); err != nil {
		logger.Fatal(fmt.Errorf("create data directory: %w", err))
	}
	database, err := store.Open(configuration.DatabasePath())
	if err != nil {
		logger.Fatal(err)
	}
	defer database.Close()
	client := &http.Client{Timeout: configuration.HTTPTimeout}
	service := delivery.Service{Store: database, Selector: selection.Selector{Random: rand.New(rand.NewSource(time.Now().UnixNano())), Hadith: hadeethenc.Client{HTTPClient: client}}, Quran: quranenc.Client{HTTPClient: client}, Hadith: hadeethenc.Client{HTTPClient: client}, Writer: os.Stdout}
	mode := "run"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch mode {
	case "once":
		if err := service.Attempt(ctx, configuration.ConsoleLanguage, time.Now(), time.Now()); err != nil {
			logger.Printf("delivery: %v", err)
		}
	case "run":
		if configuration.TelegramBotToken == "" {
			logger.Fatal("TELEGRAM_BOT_TOKEN is required for run")
		}
		telegramBot, err := telegram.New(configuration.TelegramBotToken, database, logger)
		if err != nil {
			logger.Fatal(err)
		}
		service.Broadcaster = telegramBot.Broadcast
		go telegramBot.Start(ctx)
		runSchedule(ctx, configuration, database, service, logger)
	default:
		logger.Fatalf("unknown command %q; use run or once", mode)
	}
}

func runSchedule(ctx context.Context, configuration config.Config, database *store.Store, service delivery.Service, logger *log.Logger) {
	location, _ := time.LoadLocation(configuration.Timezone)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			logger.Printf("stopping: %v", err)
			return
		}
		now := time.Now().In(location)
		if inWindow(now, configuration, location) {
			languages, err := database.SubscribedTelegramLanguages(ctx)
			if err != nil {
				logger.Printf("read subscribed languages: %v", err)
			} else if err := deliverScheduled(ctx, now, configuration, database, service, languages, logger); err != nil {
				logger.Printf("schedule: %v", err)
			}
		}
		select {
		case <-ctx.Done():
			logger.Printf("stopping: %v", ctx.Err())
			return
		case <-ticker.C:
		}
	}
}

func deliverScheduled(ctx context.Context, now time.Time, configuration config.Config, database *store.Store, service delivery.Service, languages []string, logger *log.Logger) error {
	slots, err := schedule.Slots(now, configuration.SendWindowStart, configuration.SendWindowEnd, configuration.DailyNotificationCount, now.Location())
	if err != nil {
		return err
	}
	for _, language := range languages {
		pending, err := database.HasPending(ctx, language)
		if err != nil {
			logger.Printf("read pending delivery language=%s: %v", language, err)
			continue
		}
		if pending {
			if err := service.Attempt(ctx, language, now, now); err != nil {
				logger.Printf("pending delivery language=%s: %v", language, err)
			}
			continue
		}
		for _, slot := range slots {
			if now.Before(slot) || now.Sub(slot) > 15*time.Second {
				continue
			}
			claimed, err := database.ClaimSlot(ctx, language, slot)
			if err != nil {
				logger.Printf("claim slot language=%s: %v", language, err)
				continue
			}
			if claimed {
				if err := service.Attempt(ctx, language, slot, now); err != nil {
					logger.Printf("delivery language=%s: %v", language, err)
				}
			}
		}
	}
	return nil
}

func inWindow(now time.Time, configuration config.Config, location *time.Location) bool {
	slots, err := schedule.Slots(now, configuration.SendWindowStart, configuration.SendWindowEnd, 2, location)
	return err == nil && !now.Before(slots[0]) && !now.After(slots[1])
}
