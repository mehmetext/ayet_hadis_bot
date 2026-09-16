package quranenc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "https://quranenc.com/api/v1"

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

func (client Client) FetchAyah(ctx context.Context, translationKey string, surahNumber, ayahNumber int) (Verse, error) {
	var response struct {
		Result struct {
			Surah       string `json:"sura"`
			Ayah        string `json:"aya"`
			ArabicText  string `json:"arabic_text"`
			Translation string `json:"translation"`
		} `json:"result"`
	}
	if err := client.getJSON(ctx, fmt.Sprintf("/translation/aya/%s/%d/%d", url.PathEscape(translationKey), surahNumber, ayahNumber), &response); err != nil {
		return Verse{}, err
	}
	var parsedSurah, parsedAyah int
	if _, err := fmt.Sscanf(response.Result.Surah, "%d", &parsedSurah); err != nil {
		return Verse{}, fmt.Errorf("parse surah: %w", err)
	}
	if _, err := fmt.Sscanf(response.Result.Ayah, "%d", &parsedAyah); err != nil {
		return Verse{}, fmt.Errorf("parse ayah: %w", err)
	}
	return Verse{SurahNumber: parsedSurah, AyahNumber: parsedAyah, ArabicText: response.Result.ArabicText, Translation: response.Result.Translation}, nil
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
