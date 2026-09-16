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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	verse, err := client.FetchQuranVerse(ctx, cfg.QuranVerseNumber, cfg.QuranEdition)
	if err != nil {
		logger.Fatal(err)
	}
	hadith, err := client.FetchHadith(ctx, cfg.HadithEdition, cfg.HadithNumber)
	if err != nil {
		logger.Fatal(err)
	}
	if err := output.PrintContent(os.Stdout, verse, hadith); err != nil {
		logger.Fatal(err)
	}

	logger.Printf("content interval configured as %s", cfg.ContentInterval.Round(time.Second))
}
