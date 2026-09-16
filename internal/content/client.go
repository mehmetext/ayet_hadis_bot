package content

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	quranBaseURL  = "https://api.alquran.cloud/v1/ayah"
	hadithBaseURL = "https://cdn.jsdelivr.net/gh/fawazahmed0/hadith-api@1/editions"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

func (c *Client) FetchQuranVerse(ctx context.Context, number int, edition string) (Verse, error) {
	body, err := c.get(ctx, fmt.Sprintf("%s/%d/%s", quranBaseURL, number, edition))
	if err != nil {
		return Verse{}, err
	}
	return DecodeQuranResponse(body)
}

func (c *Client) FetchHadith(ctx context.Context, edition string, number int) (Hadith, error) {
	body, err := c.get(ctx, fmt.Sprintf("%s/%s/%d.json", hadithBaseURL, edition, number))
	if err != nil {
		return Hadith{}, err
	}
	return DecodeHadithResponse(body)
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request %s: unexpected status %s", url, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}
