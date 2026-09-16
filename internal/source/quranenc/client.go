package quranenc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "https://quranenc.com/api/v1"

type Translation struct {
	Key     string `json:"key"`
	Version string `json:"version"`
	Title   string `json:"title"`
}
type Verse struct {
	SurahNumber int
	AyahNumber  int
	ArabicText  string
	Translation string
}
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func (client Client) AvailableTranslations(ctx context.Context, languageCode string) ([]Translation, error) {
	var response struct {
		Translations []Translation `json:"translations"`
	}
	if err := client.getJSON(ctx, "/translations/list/"+url.PathEscape(languageCode)+"/", &response); err != nil {
		return nil, err
	}
	return response.Translations, nil
}

func (client Client) FetchSurah(ctx context.Context, translationKey string, surahNumber int) ([]Verse, error) {
	var response struct {
		Result []struct {
			Surah       string `json:"sura"`
			Ayah        string `json:"aya"`
			ArabicText  string `json:"arabic_text"`
			Translation string `json:"translation"`
		} `json:"result"`
	}
	if err := client.getJSON(ctx, fmt.Sprintf("/translation/sura/%s/%d", url.PathEscape(translationKey), surahNumber), &response); err != nil {
		return nil, err
	}
	verses := make([]Verse, 0, len(response.Result))
	for _, item := range response.Result {
		var parsedSurah, parsedAyah int
		if _, err := fmt.Sscanf(item.Surah, "%d", &parsedSurah); err != nil {
			return nil, fmt.Errorf("parse surah: %w", err)
		}
		if _, err := fmt.Sscanf(item.Ayah, "%d", &parsedAyah); err != nil {
			return nil, fmt.Errorf("parse ayah: %w", err)
		}
		verses = append(verses, Verse{SurahNumber: parsedSurah, AyahNumber: parsedAyah, ArabicText: item.ArabicText, Translation: item.Translation})
	}
	return verses, nil
}

func (client Client) getJSON(ctx context.Context, path string, target any) error {
	base := client.BaseURL
	if base == "" {
		base = baseURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return fmt.Errorf("create QuranEnc request: %w", err)
	}
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("request QuranEnc: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("QuranEnc status: %s", response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode QuranEnc response: %w", err)
	}
	return nil
}
