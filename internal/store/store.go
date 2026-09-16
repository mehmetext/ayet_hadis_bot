package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schemaVersion = "2"

type ContentType string

const (
	ContentTypeVerse  ContentType = "verse"
	ContentTypeHadith ContentType = "hadith"
)

type Candidate struct {
	LanguageCode string
	Type         ContentType
	Source       string
	ID           string
}
type Pending struct {
	Candidate     Candidate
	Slot          time.Time
	Attempts      int
	NextAttemptAt time.Time
}

var ErrUnavailable = errors.New("candidate unavailable")

type Store struct{ database *sql.DB }

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
func (store *Store) Close() error { return store.database.Close() }

func (store *Store) NextType(ctx context.Context, languageCode string) (ContentType, error) {
	var value string
	err := store.database.QueryRowContext(ctx, `SELECT next_content_type FROM delivery_state WHERE language_code = ?`, languageCode).Scan(&value)
	if err == sql.ErrNoRows {
		return ContentTypeVerse, nil
	}
	if err != nil {
		return "", fmt.Errorf("read delivery state: %w", err)
	}
	return ContentType(value), nil
}

func (store *Store) Reserve(ctx context.Context, candidate Candidate, expiresAt time.Time) (bool, error) {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin reservation: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `DELETE FROM delivery_reservations WHERE expires_at <= ?`, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return false, fmt.Errorf("clear expired reservations: %w", err)
	}
	cycle, err := currentCycle(ctx, transaction, candidate.LanguageCode, candidate.Type)
	if err != nil {
		return false, err
	}
	var count int
	err = transaction.QueryRowContext(ctx, `SELECT COUNT(*) FROM delivery_history WHERE language_code=? AND content_type=? AND source=? AND content_id=? AND cycle=?`, candidate.LanguageCode, candidate.Type, candidate.Source, candidate.ID, cycle).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check delivery history: %w", err)
	}
	if count > 0 {
		return false, nil
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO delivery_reservations(language_code, content_type, source, content_id, expires_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(language_code, content_type, source, content_id) DO NOTHING`, candidate.LanguageCode, candidate.Type, candidate.Source, candidate.ID, expiresAt.UTC().Format(time.RFC3339))
	if err != nil {
		return false, fmt.Errorf("reserve candidate: %w", err)
	}
	result, err := transaction.ExecContext(ctx, `SELECT changes()`)
	if err != nil {
		return false, fmt.Errorf("read reservation result: %w", err)
	}
	changed, _ := result.RowsAffected()
	if err := transaction.Commit(); err != nil {
		return false, fmt.Errorf("commit reservation: %w", err)
	}
	return changed == 1, nil
}

func (store *Store) Complete(ctx context.Context, candidate Candidate, slot time.Time) error {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin completion: %w", err)
	}
	defer transaction.Rollback()
	cycle, err := currentCycle(ctx, transaction, candidate.LanguageCode, candidate.Type)
	if err != nil {
		return err
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO delivery_history(language_code, content_type, source, content_id, cycle, delivered_at) VALUES (?, ?, ?, ?, ?, ?)`, candidate.LanguageCode, candidate.Type, candidate.Source, candidate.ID, cycle, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("record delivery history: %w", err)
	}
	next := ContentTypeVerse
	if candidate.Type == ContentTypeVerse {
		next = ContentTypeHadith
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO delivery_state(language_code, next_content_type, last_successful_slot) VALUES (?, ?, ?) ON CONFLICT(language_code) DO UPDATE SET next_content_type=excluded.next_content_type, last_successful_slot=excluded.last_successful_slot`, candidate.LanguageCode, next, slot.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("advance delivery state: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM delivery_reservations WHERE language_code=? AND content_type=? AND source=? AND content_id=?`, candidate.LanguageCode, candidate.Type, candidate.Source, candidate.ID); err != nil {
		return fmt.Errorf("clear reservation: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM pending_deliveries WHERE language_code=?`, candidate.LanguageCode); err != nil {
		return fmt.Errorf("clear pending delivery: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit delivery completion: %w", err)
	}
	return nil
}

func (store *Store) Release(ctx context.Context, candidate Candidate) error {
	_, err := store.database.ExecContext(ctx, `DELETE FROM delivery_reservations WHERE language_code=? AND content_type=? AND source=? AND content_id=?`, candidate.LanguageCode, candidate.Type, candidate.Source, candidate.ID)
	if err != nil {
		return fmt.Errorf("release reservation: %w", err)
	}
	return nil
}
func (store *Store) SavePending(ctx context.Context, pending Pending) error {
	_, err := store.database.ExecContext(ctx, `INSERT INTO pending_deliveries(language_code, content_type, source, content_id, slot_at, attempts, next_attempt_at) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(language_code) DO UPDATE SET content_type=excluded.content_type, source=excluded.source, content_id=excluded.content_id, slot_at=excluded.slot_at, attempts=excluded.attempts, next_attempt_at=excluded.next_attempt_at`, pending.Candidate.LanguageCode, pending.Candidate.Type, pending.Candidate.Source, pending.Candidate.ID, pending.Slot.UTC().Format(time.RFC3339), pending.Attempts, pending.NextAttemptAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save pending delivery: %w", err)
	}
	return nil
}
func (store *Store) Pending(ctx context.Context, languageCode string) (Pending, bool, error) {
	var pending Pending
	var slot, next string
	err := store.database.QueryRowContext(ctx, `SELECT content_type, source, content_id, slot_at, attempts, next_attempt_at FROM pending_deliveries WHERE language_code=?`, languageCode).Scan(&pending.Candidate.Type, &pending.Candidate.Source, &pending.Candidate.ID, &slot, &pending.Attempts, &next)
	if err == sql.ErrNoRows {
		return Pending{}, false, nil
	}
	if err != nil {
		return Pending{}, false, fmt.Errorf("read pending delivery: %w", err)
	}
	pending.Candidate.LanguageCode = languageCode
	pending.Slot, _ = time.Parse(time.RFC3339, slot)
	pending.NextAttemptAt, _ = time.Parse(time.RFC3339, next)
	return pending, true, nil
}
func (store *Store) PendingDue(ctx context.Context, languageCode string, now time.Time) (Pending, bool, error) {
	pending, exists, err := store.Pending(ctx, languageCode)
	return pending, exists && !pending.NextAttemptAt.After(now), err
}
func (store *Store) HasPending(ctx context.Context, languageCode string) (bool, error) {
	var count int
	err := store.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM pending_deliveries WHERE language_code=?`, languageCode).Scan(&count)
	return count > 0, err
}
func (store *Store) ClaimSlot(ctx context.Context, languageCode string, slot time.Time) (bool, error) {
	result, err := store.database.ExecContext(ctx, `INSERT INTO delivery_slots(language_code, slot_at) VALUES (?, ?) ON CONFLICT(language_code, slot_at) DO NOTHING`, languageCode, slot.UTC().Format(time.RFC3339))
	if err != nil {
		return false, fmt.Errorf("claim delivery slot: %w", err)
	}
	changed, err := result.RowsAffected()
	return changed == 1, err
}
func (store *Store) ClearCycle(ctx context.Context, languageCode string, contentType ContentType) error {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	cycle, err := currentCycle(ctx, transaction, languageCode, contentType)
	if err != nil {
		return err
	}
	_, err = transaction.ExecContext(ctx, `DELETE FROM delivery_history WHERE language_code=? AND content_type=? AND cycle=?`, languageCode, contentType, cycle)
	if err != nil {
		return err
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO delivery_cycles(language_code, content_type, cycle) VALUES (?, ?, ?) ON CONFLICT(language_code, content_type) DO UPDATE SET cycle=excluded.cycle`, languageCode, contentType, cycle+1)
	if err != nil {
		return err
	}
	return transaction.Commit()
}
func (store *Store) HistoryCount(ctx context.Context, languageCode string, contentType ContentType) (int, error) {
	var count int
	err := store.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM delivery_history WHERE language_code=? AND content_type=?`, languageCode, contentType).Scan(&count)
	return count, err
}

func currentCycle(ctx context.Context, transaction *sql.Tx, languageCode string, contentType ContentType) (int, error) {
	var cycle int
	err := transaction.QueryRowContext(ctx, `SELECT cycle FROM delivery_cycles WHERE language_code=? AND content_type=?`, languageCode, contentType).Scan(&cycle)
	if err == sql.ErrNoRows {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read delivery cycle: %w", err)
	}
	return cycle, nil
}
func (store *Store) migrate(ctx context.Context) error {
	if _, err := store.database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_meta(key TEXT PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		return err
	}
	var version string
	err := store.database.QueryRowContext(ctx, `SELECT value FROM schema_meta WHERE key='schema_version'`).Scan(&version)
	if err == sql.ErrNoRows {
		for _, name := range []string{"sync_state", "quran_editions", "quran_verses", "hadith_editions", "hadiths", "delivery_state", "delivery_history"} {
			if _, err := store.database.ExecContext(ctx, `DROP TABLE IF EXISTS `+name); err != nil {
				return err
			}
		}
		_, err = store.database.ExecContext(ctx, `INSERT INTO schema_meta(key,value) VALUES('schema_version',?)`, schemaVersion)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if version != schemaVersion {
		return fmt.Errorf("unsupported schema version: %s", version)
	}
	statements := []string{`CREATE TABLE IF NOT EXISTS delivery_state(language_code TEXT PRIMARY KEY, next_content_type TEXT NOT NULL, last_successful_slot TEXT)`, `CREATE TABLE IF NOT EXISTS delivery_cycles(language_code TEXT NOT NULL, content_type TEXT NOT NULL, cycle INTEGER NOT NULL, PRIMARY KEY(language_code, content_type))`, `CREATE TABLE IF NOT EXISTS delivery_history(language_code TEXT NOT NULL, content_type TEXT NOT NULL, source TEXT NOT NULL, content_id TEXT NOT NULL, cycle INTEGER NOT NULL, delivered_at TEXT NOT NULL, PRIMARY KEY(language_code, content_type, source, content_id, cycle))`, `CREATE TABLE IF NOT EXISTS delivery_reservations(language_code TEXT NOT NULL, content_type TEXT NOT NULL, source TEXT NOT NULL, content_id TEXT NOT NULL, expires_at TEXT NOT NULL, PRIMARY KEY(language_code, content_type, source, content_id))`, `CREATE TABLE IF NOT EXISTS pending_deliveries(language_code TEXT PRIMARY KEY, content_type TEXT NOT NULL, source TEXT NOT NULL, content_id TEXT NOT NULL, slot_at TEXT NOT NULL, attempts INTEGER NOT NULL, next_attempt_at TEXT NOT NULL)`, `CREATE TABLE IF NOT EXISTS delivery_slots(language_code TEXT NOT NULL, slot_at TEXT NOT NULL, PRIMARY KEY(language_code, slot_at))`}
	for _, statement := range statements {
		if _, err := store.database.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}
	return nil
}
