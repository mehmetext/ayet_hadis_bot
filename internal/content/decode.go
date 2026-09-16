package content

import (
	"encoding/json"
	"fmt"
)

type quranResponse struct {
	Data quranData `json:"data"`
}

type quranData struct {
	Number        int    `json:"number"`
	Text          string `json:"text"`
	NumberInSurah int    `json:"numberInSurah"`
	Surah         struct {
		EnglishName string `json:"englishName"`
	} `json:"surah"`
}

type hadithResponse struct {
	Hadiths []struct {
		HadithNumber int    `json:"hadithnumber"`
		Text         string `json:"text"`
		Reference    struct {
			Book   json.RawMessage `json:"book"`
			Hadith int    `json:"hadith"`
		} `json:"reference"`
		Grades []struct {
			Grade string `json:"grade"`
		} `json:"grades"`
	} `json:"hadiths"`
}

func DecodeQuranResponse(input []byte) (Verse, error) {
	var response quranResponse
	if err := json.Unmarshal(input, &response); err != nil {
		return Verse{}, fmt.Errorf("decode Quran response: %w", err)
	}
	if response.Data.Text == "" || response.Data.Surah.EnglishName == "" {
		return Verse{}, fmt.Errorf("decode Quran response: missing verse data")
	}
	return Verse{
		Number:        response.Data.Number,
		SurahName:     response.Data.Surah.EnglishName,
		NumberInSurah: response.Data.NumberInSurah,
		Text:          response.Data.Text,
		Reference:     fmt.Sprintf("%s %d:%d", response.Data.Surah.EnglishName, response.Data.NumberInSurah, response.Data.Number),
	}, nil
}

func DecodeHadithResponse(input []byte) (Hadith, error) {
	var response hadithResponse
	if err := json.Unmarshal(input, &response); err != nil {
		return Hadith{}, fmt.Errorf("decode hadith response: %w", err)
	}
	if len(response.Hadiths) == 0 || response.Hadiths[0].Text == "" {
		return Hadith{}, fmt.Errorf("decode hadith response: no hadith data")
	}
	item := response.Hadiths[0]
	book := referenceValue(item.Reference.Book)
	if book == "" {
		return Hadith{}, fmt.Errorf("decode hadith response: missing book reference")
	}
	grade := ""
	if len(item.Grades) > 0 {
		grade = item.Grades[0].Grade
	}
	return Hadith{
		Number:    item.HadithNumber,
		Text:      item.Text,
		Grade:     grade,
		Reference: fmt.Sprintf("Book %s, Hadith %d", book, item.Reference.Hadith),
	}, nil
}

func referenceValue(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var number json.Number
	if json.Unmarshal(raw, &number) == nil {
		return number.String()
	}
	return ""
}
