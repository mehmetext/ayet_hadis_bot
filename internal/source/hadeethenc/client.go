package hadeethenc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "https://hadeethenc.com/api/v1"

type Category struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
}
type Listing struct {
	ID string `json:"id"`
}
type Hadith struct {
	ID          string `json:"id"`
	Hadeeth     string `json:"hadeeth"`
	Attribution string `json:"attribution"`
	Grade       string `json:"grade"`
	Explanation string `json:"explanation"`
}
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func (client Client) RootCategories(ctx context.Context, languageCode string) ([]Category, error) {
	var categories []Category
	err := client.getJSON(ctx, "/categories/roots/?language="+url.QueryEscape(languageCode), &categories)
	return categories, err
}
func (client Client) List(ctx context.Context, languageCode, categoryID string, page, perPage int) ([]Listing, int, error) {
	var response struct {
		Data []Listing `json:"data"`
		Meta struct {
			LastPage int `json:"last_page"`
		} `json:"meta"`
	}
	path := fmt.Sprintf("/hadeeths/list/?language=%s&category_id=%s&page=%d&per_page=%d", url.QueryEscape(languageCode), url.QueryEscape(categoryID), page, perPage)
	if err := client.getJSON(ctx, path, &response); err != nil {
		return nil, 0, err
	}
	return response.Data, response.Meta.LastPage, nil
}
func (client Client) One(ctx context.Context, languageCode, hadithID string) (Hadith, error) {
	var hadith Hadith
	err := client.getJSON(ctx, "/hadeeths/one/?language="+url.QueryEscape(languageCode)+"&id="+url.QueryEscape(hadithID), &hadith)
	return hadith, err
}
func (client Client) getJSON(ctx context.Context, path string, target any) error {
	base := client.BaseURL
	if base == "" {
		base = baseURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return fmt.Errorf("create HadeethEnc request: %w", err)
	}
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("request HadeethEnc: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("HadeethEnc status: %s", response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode HadeethEnc response: %w", err)
	}
	return nil
}
