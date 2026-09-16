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

	"github.com/example/ayet-hadis-bot/internal/config"
	"github.com/example/ayet-hadis-bot/internal/schedule"
	"github.com/example/ayet-hadis-bot/internal/source/hadeethenc"
	"github.com/example/ayet-hadis-bot/internal/source/quranenc"
	"github.com/example/ayet-hadis-bot/internal/store"
	"github.com/example/ayet-hadis-bot/internal/sync"
)

const defaultLanguage = "tur"

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
	synchronizer := sync.Service{Quran: sync.QuranSynchronizer{Client: quranenc.Client{HTTPClient: client}, Store: database}, Hadith: sync.HadithSynchronizer{Client: hadeethenc.Client{HTTPClient: client}, Store: database}, Logger: logger}
	mode := "run"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch mode {
	case "sync":
		synchronizer.SyncAll(ctx)
	case "once":
		printNext(ctx, database, logger)
	case "run":
		synchronizer.SyncAll(ctx)
		runSchedule(ctx, configuration, database, logger)
	default:
		logger.Fatalf("unknown command %q; use run, sync, or once", mode)
	}
}

func runSchedule(ctx context.Context, configuration config.Config, database *store.Store, logger *log.Logger) {
	location, _ := time.LoadLocation(configuration.Timezone)
	for {
		if err := ctx.Err(); err != nil {
			logger.Printf("stopping: %v", err)
			return
		}
		now := time.Now().In(location)
		slots, err := schedule.Slots(now, configuration.SendWindowStart, configuration.SendWindowEnd, configuration.DailyNotificationCount, location)
		if err != nil {
			logger.Printf("schedule: %v", err)
			return
		}
		for _, slot := range slots {
			if slot.Before(now) {
				continue
			}
			if wait := time.Until(slot); wait > 0 {
				if !waitFor(ctx, wait) {
					logger.Printf("stopping: %v", ctx.Err())
					return
				}
			}
			printNext(ctx, database, logger)
		}
		if !waitFor(ctx, time.Until(time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, location))) {
			logger.Printf("stopping: %v", ctx.Err())
			return
		}
	}
}

func waitFor(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func printNext(ctx context.Context, database *store.Store, logger *log.Logger) {
	contentType, err := database.NextContentType(ctx)
	if err != nil {
		logger.Printf("choose content type: %v", err)
		return
	}
	item, err := database.SelectUndelivered(ctx, defaultLanguage, contentType)
	if err != nil {
		logger.Printf("select %s: %v", contentType, err)
		return
	}
	if item.Type == store.ContentTypeVerse {
		fmt.Printf("AYET — %s\n%s\n%s\nKaynak: %s\n", item.Reference, item.ArabicText, item.Text, item.Attribution)
		return
	}
	fmt.Printf("HADİS — %s\n%s\nDerece: %s\nAçıklama: %s\nKaynak: %s\n", item.Reference, item.Text, item.Grade, item.Explanation, item.Attribution)
}
