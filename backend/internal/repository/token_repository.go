package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gant123/jobTracker/internal/crypto"
	"golang.org/x/oauth2"
)

// Interface the handler expects
type TokenRepository interface {
	Save(ctx context.Context, userID int, provider string, tok *oauth2.Token) error
	Get(ctx context.Context, userID int, provider string) (*oauth2.Token, error)
	Delete(ctx context.Context, userID int, provider string) error
}

type PostgresTokenRepository struct {
	db    *sql.DB
	box   *crypto.SecretBox
	nowFn func() time.Time
}

func NewPostgresTokenRepository(db *sql.DB, box *crypto.SecretBox) *PostgresTokenRepository {
	return &PostgresTokenRepository{
		db:    db,
		box:   box,
		nowFn: func() time.Time { return time.Now().UTC() },
	}
}

func (r *PostgresTokenRepository) Save(ctx context.Context, userID int, provider string, tok *oauth2.Token) error {
	if tok == nil {
		return errors.New("nil token")
	}
	encAccess, err := r.box.Seal([]byte(tok.AccessToken))
	if err != nil {
		return err
	}
	var encRefresh []byte
	if tok.RefreshToken != "" {
		encRefresh, err = r.box.Seal([]byte(tok.RefreshToken))
		if err != nil {
			return err
		}
	}

	// upsert
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO email_tokens (user_id, provider, access_token_enc, refresh_token_enc, expiry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id, provider)
		DO UPDATE SET access_token_enc=EXCLUDED.access_token_enc,
		              refresh_token_enc=EXCLUDED.refresh_token_enc,
		              expiry=EXCLUDED.expiry,
		              updated_at=NOW()
	`, userID, provider, encAccess, encRefresh, nullableTime(tok.Expiry))
	return err
}

func (r *PostgresTokenRepository) Get(ctx context.Context, userID int, provider string) (*oauth2.Token, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT access_token_enc, COALESCE(refresh_token_enc, ''), expiry
		FROM email_tokens
		WHERE user_id=$1 AND provider=$2
	`, userID, provider)

	var encAccess, encRefresh []byte
	var expiry sql.NullTime
	if err := row.Scan(&encAccess, &encRefresh, &expiry); err != nil {
		return nil, err
	}

	at, err := r.box.Open(encAccess)
	if err != nil {
		return nil, err
	}
	var rt string
	if len(encRefresh) > 0 {
		pt, err := r.box.Open(encRefresh)
		if err != nil {
			return nil, err
		}
		rt = string(pt)
	}

	tok := &oauth2.Token{
		AccessToken:  string(at),
		RefreshToken: rt,
		TokenType:    "Bearer",
	}
	if expiry.Valid {
		tok.Expiry = expiry.Time
	}
	return tok, nil
}

func (r *PostgresTokenRepository) Delete(ctx context.Context, userID int, provider string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM email_tokens WHERE user_id=$1 AND provider=$2
	`, userID, provider)
	return err
}

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
