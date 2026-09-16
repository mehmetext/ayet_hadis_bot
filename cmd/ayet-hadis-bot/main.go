package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/ayet-hadis-bot/internal/config"
	"github.com/example/ayet-hadis-bot/internal/content"
	"github.com/example/ayet-hadis-bot/internal/output"
)

func main() {
	logger := log.New(os.Stderr, "ayet-hadis-bot: ", log.LstdFlags)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}

	client := content.NewClient(&http.Client{Timeout: cfg.HTTPTimeout})
	ctx := context.Background()
	printContent(ctx, client, cfg, logger)
	logger.Printf("content interval configured as %s", cfg.ContentInterval.Round(time.Second))

	ticker := time.NewTicker(cfg.ContentInterval)
	defer ticker.Stop()
	for range ticker.C {
		printContent(ctx, client, cfg, logger)
	}
}

func printContent(ctx context.Context, client *content.Client, cfg config.Config, logger *log.Logger) {
	requestContext, cancel := context.WithTimeout(ctx, cfg.HTTPTimeout)
	defer cancel()

	verse, err := client.FetchQuranVerse(requestContext, cfg.QuranVerseNumber, cfg.QuranEdition)
	if err != nil {
		logger.Printf("fetch Quran verse: %v", err)
		return
	}
	hadith, err := client.FetchHadith(requestContext, cfg.HadithEdition, cfg.HadithNumber)
	if err != nil {
		logger.Printf("fetch hadith: %v", err)
		return
	}
	if err := output.PrintContent(os.Stdout, verse, hadith); err != nil {
		logger.Printf("print content: %v", err)
	}
}
