package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type ContentType string

const (
	ContentTypeVerse  ContentType = "verse"
	ContentTypeHadith ContentType = "hadith"
)

type Store struct {
	database *sql.DB
}

type QuranVerse struct {
	SurahNumber int
	AyahNumber  int
	ArabicText  string
	Text        string
}

type Hadith struct {
	ID             string
	CollectionName string
	Reference      string
	Text           string
	Grade          string
	Explanation    string
}

type DeliveredContent struct {
	Type        ContentType
	Language    string
	ArabicText  string
	Text        string
	Reference   string
	Grade       string
	Explanation string
	Attribution string
}

func Open(path string) (*Store, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	database.SetMaxOpenConns(1)
	store := &Store{database: database}
	if err := store.migrate(context.Background()); err != nil {
		_ = database.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.database.Close()
}

func (s *Store) IsSyncComplete(ctx context.Context, source, languageCode, editionKey string) (bool, error) {
	var status string
	err := s.database.QueryRowContext(ctx, `SELECT status FROM sync_state WHERE source = ? AND language_code = ? AND edition_key = ?`, source, languageCode, editionKey).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read sync state: %w", err)
	}
	return status == "completed", nil
}

func (s *Store) NextContentType(ctx context.Context) (ContentType, error) {
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin delivery state transaction: %w", err)
	}
	defer transaction.Rollback()

	var next string
	err = transaction.QueryRowContext(ctx, `SELECT value FROM delivery_state WHERE key = 'next_content_type'`).Scan(&next)
	if err == sql.ErrNoRows {
		next = string(ContentTypeVerse)
	} else if err != nil {
		return "", fmt.Errorf("read delivery state: %w", err)
	}

	contentType := ContentType(next)
	if contentType != ContentTypeVerse && contentType != ContentTypeHadith {
		return "", fmt.Errorf("invalid content type in delivery state: %q", contentType)
	}
	following := ContentTypeVerse
	if contentType == ContentTypeVerse {
		following = ContentTypeHadith
	}
	_, err = transaction.ExecContext(ctx, `
		INSERT INTO delivery_state(key, value) VALUES ('next_content_type', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, following)
	if err != nil {
		return "", fmt.Errorf("write delivery state: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return "", fmt.Errorf("commit delivery state: %w", err)
	}
	return contentType, nil
}

func (s *Store) ReplaceQuranEdition(ctx context.Context, languageCode, editionKey, publisher, attribution, version string, verses []QuranVerse) error {
	if len(verses) != 6236 {
		return fmt.Errorf("Quran edition %q has %d verses; want 6236", editionKey, len(verses))
	}
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Quran import: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `DELETE FROM quran_editions WHERE edition_key = ?`, editionKey); err != nil {
		return fmt.Errorf("clear Quran edition: %w", err)
	}
	result, err := transaction.ExecContext(ctx, `INSERT INTO quran_editions(language_code, edition_key, publisher, attribution, source_version) VALUES (?, ?, ?, ?, ?)`, languageCode, editionKey, publisher, attribution, version)
	if err != nil {
		return fmt.Errorf("insert Quran edition: %w", err)
	}
	editionID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read Quran edition ID: %w", err)
	}
	statement, err := transaction.PrepareContext(ctx, `INSERT INTO quran_verses(edition_id, surah_number, ayah_number, arabic_text, text) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare Quran verse import: %w", err)
	}
	defer statement.Close()
	for _, verse := range verses {
		if _, err := statement.ExecContext(ctx, editionID, verse.SurahNumber, verse.AyahNumber, verse.ArabicText, verse.Text); err != nil {
			return fmt.Errorf("insert Quran verse: %w", err)
		}
	}
	return commitSyncState(ctx, transaction, "quranenc", languageCode, editionKey, len(verses), version)
}

func (s *Store) ReplaceHadithEdition(ctx context.Context, languageCode, editionKey, publisher, attribution, version string, hadiths []Hadith) error {
	if len(hadiths) == 0 {
		return fmt.Errorf("hadith edition %q has no records", editionKey)
	}
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin hadith import: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `DELETE FROM hadith_editions WHERE edition_key = ?`, editionKey); err != nil {
		return fmt.Errorf("clear hadith edition: %w", err)
	}
	result, err := transaction.ExecContext(ctx, `INSERT INTO hadith_editions(language_code, edition_key, publisher, attribution, source_version) VALUES (?, ?, ?, ?, ?)`, languageCode, editionKey, publisher, attribution, version)
	if err != nil {
		return fmt.Errorf("insert hadith edition: %w", err)
	}
	editionID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read hadith edition ID: %w", err)
	}
	statement, err := transaction.PrepareContext(ctx, `INSERT INTO hadiths(edition_id, source_id, collection_name, reference, text, grade, explanation) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare hadith import: %w", err)
	}
	defer statement.Close()
	for _, hadith := range hadiths {
		if _, err := statement.ExecContext(ctx, editionID, hadith.ID, hadith.CollectionName, hadith.Reference, hadith.Text, hadith.Grade, hadith.Explanation); err != nil {
			return fmt.Errorf("insert hadith: %w", err)
		}
	}
	return commitSyncState(ctx, transaction, "hadeethenc", languageCode, editionKey, len(hadiths), version)
}

func (s *Store) SelectUndelivered(ctx context.Context, languageCode string, contentType ContentType) (DeliveredContent, error) {
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return DeliveredContent{}, fmt.Errorf("begin content selection: %w", err)
	}
	defer transaction.Rollback()
	content, contentID, err := selectUndelivered(ctx, transaction, languageCode, contentType)
	if err == sql.ErrNoRows {
		if _, clearErr := transaction.ExecContext(ctx, `DELETE FROM delivery_history WHERE language_code = ? AND content_type = ?`, languageCode, contentType); clearErr != nil {
			return DeliveredContent{}, fmt.Errorf("reset delivery history: %w", clearErr)
		}
		content, contentID, err = selectUndelivered(ctx, transaction, languageCode, contentType)
	}
	if err != nil {
		return DeliveredContent{}, fmt.Errorf("select local %s: %w", contentType, err)
	}
	if _, err := transaction.ExecContext(ctx, `INSERT INTO delivery_history(content_type, language_code, content_id, delivered_at) VALUES (?, ?, ?, ?)`, contentType, languageCode, contentID, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return DeliveredContent{}, fmt.Errorf("record delivery history: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return DeliveredContent{}, fmt.Errorf("commit content selection: %w", err)
	}
	return content, nil
}

func commitSyncState(ctx context.Context, transaction *sql.Tx, source, languageCode, editionKey string, count int, version string) error {
	_, err := transaction.ExecContext(ctx, `INSERT INTO sync_state(source, language_code, edition_key, status, expected_count, imported_count, source_version, completed_at) VALUES (?, ?, ?, 'completed', ?, ?, ?, ?) ON CONFLICT(source, language_code, edition_key) DO UPDATE SET status = 'completed', expected_count = excluded.expected_count, imported_count = excluded.imported_count, source_version = excluded.source_version, last_error = '', completed_at = excluded.completed_at`, source, languageCode, editionKey, count, count, version, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("write sync state: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	return nil
}

func selectUndelivered(ctx context.Context, transaction *sql.Tx, languageCode string, contentType ContentType) (DeliveredContent, string, error) {
	if contentType == ContentTypeVerse {
		var content DeliveredContent
		var id string
		err := transaction.QueryRowContext(ctx, `SELECT v.edition_id || ':' || v.surah_number || ':' || v.ayah_number, v.arabic_text, v.text, v.surah_number || ':' || v.ayah_number, e.attribution FROM quran_verses v JOIN quran_editions e ON e.id = v.edition_id WHERE e.language_code = ? AND NOT EXISTS (SELECT 1 FROM delivery_history h WHERE h.content_type = 'verse' AND h.language_code = ? AND h.content_id = v.edition_id || ':' || v.surah_number || ':' || v.ayah_number) ORDER BY RANDOM() LIMIT 1`, languageCode, languageCode).Scan(&id, &content.ArabicText, &content.Text, &content.Reference, &content.Attribution)
		content.Type, content.Language = ContentTypeVerse, languageCode
		return content, id, err
	}
	var content DeliveredContent
	var id string
	err := transaction.QueryRowContext(ctx, `SELECT h.edition_id || ':' || h.id, h.text, h.reference, h.grade, h.explanation, e.attribution FROM hadiths h JOIN hadith_editions e ON e.id = h.edition_id WHERE e.language_code = ? AND NOT EXISTS (SELECT 1 FROM delivery_history x WHERE x.content_type = 'hadith' AND x.language_code = ? AND x.content_id = h.edition_id || ':' || h.id) ORDER BY RANDOM() LIMIT 1`, languageCode, languageCode).Scan(&id, &content.Text, &content.Reference, &content.Grade, &content.Explanation, &content.Attribution)
	content.Type, content.Language = ContentTypeHadith, languageCode
	return content, id, err
}

func (s *Store) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sync_state (
			source TEXT NOT NULL,
			language_code TEXT NOT NULL,
			edition_key TEXT NOT NULL,
			status TEXT NOT NULL,
			expected_count INTEGER NOT NULL DEFAULT 0,
			imported_count INTEGER NOT NULL DEFAULT 0,
			source_version TEXT NOT NULL DEFAULT '',
			last_error TEXT NOT NULL DEFAULT '',
			completed_at TEXT,
			PRIMARY KEY(source, language_code, edition_key)
		)`,
		`CREATE TABLE IF NOT EXISTS quran_editions (
			id INTEGER PRIMARY KEY,
			language_code TEXT NOT NULL,
			edition_key TEXT NOT NULL UNIQUE,
			publisher TEXT NOT NULL,
			attribution TEXT NOT NULL,
			source_version TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS quran_verses (
			edition_id INTEGER NOT NULL REFERENCES quran_editions(id),
			surah_number INTEGER NOT NULL,
			ayah_number INTEGER NOT NULL,
			arabic_text TEXT NOT NULL,
			text TEXT NOT NULL,
			PRIMARY KEY(edition_id, surah_number, ayah_number)
		)`,
		`CREATE TABLE IF NOT EXISTS hadith_editions (
			id INTEGER PRIMARY KEY,
			language_code TEXT NOT NULL,
			edition_key TEXT NOT NULL UNIQUE,
			publisher TEXT NOT NULL,
			attribution TEXT NOT NULL,
			source_version TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS hadiths (
			id INTEGER PRIMARY KEY,
			edition_id INTEGER NOT NULL REFERENCES hadith_editions(id),
			source_id TEXT NOT NULL,
			collection_name TEXT NOT NULL,
			reference TEXT NOT NULL,
			text TEXT NOT NULL,
			grade TEXT NOT NULL DEFAULT '',
			explanation TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS delivery_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS delivery_history (
			content_type TEXT NOT NULL,
			language_code TEXT NOT NULL,
			content_id TEXT NOT NULL,
			delivered_at TEXT NOT NULL,
			PRIMARY KEY(content_type, language_code, content_id)
		)`,
	}
	for _, statement := range statements {
		if _, err := s.database.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}
	return nil
}
