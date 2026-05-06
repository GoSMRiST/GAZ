package repository

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"auth_users/internal/core/auth/dto"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type DataBase struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func InitDataBase(log *slog.Logger, ctx context.Context, connString string) (*DataBase, error) {

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DataBase{
		log:  log,
		pool: pool,
	}, nil
}

func (d *DataBase) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *DataBase) Close() {
	d.pool.Close()
}

func (d *DataBase) GetUserByEmail(ctx context.Context, email string) (*dto.AuthUser, error) {
	data := &dto.AuthUser{}

	strQuery := `
		SELECT id, email, password_hash, is_verified
		FROM users
		WHERE email = $1;
	`

	err := d.pool.QueryRow(ctx, strQuery, email).Scan(
		&data.ID,
		&data.Email,
		&data.PasswordHash,
		&data.IsVerified,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return data, nil
}

func (d *DataBase) GetUserByID(ctx context.Context, id int) (*dto.GetByIdResponse, error) {
	data := &dto.GetByIdResponse{}

	strQuery := `
		SELECT
			id,
			nickname,
			birth_date,
			gender,
			email,
			avatar_url,
			meetings_count,
			is_verified,
			subscription,
			COALESCE(bio, '')
		FROM users
		WHERE id = $1;
	`

	err := d.pool.QueryRow(ctx, strQuery, id).Scan(
		&data.ID,
		&data.Nickname,
		&data.BirthDate,
		&data.Gender,
		&data.Email,
		&data.AvatarURL,
		&data.MeetingsCount,
		&data.IsVerified,
		&data.Subscription,
		&data.Bio,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return data, nil
}

func (d *DataBase) UpdateNickname(ctx context.Context, id int, nickname *dto.UpdateNicknameRequest) error {
	strQuery := `
		UPDATE users
		SET nickname = $1
		WHERE id = $2;
	`

	result, err := d.pool.Exec(ctx, strQuery, nickname.NewNickname, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}

	return nil
}

func (d *DataBase) UpdateAvatarURL(ctx context.Context, id int, url *dto.UpdateAvatarRequest) error {
	strQuery := `
		UPDATE users
		SET avatar_url = $1
		WHERE id = $2;
	`

	result, err := d.pool.Exec(ctx, strQuery, url.NewAvatar, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}

	return nil
}

func (d *DataBase) UpdatePassword(ctx context.Context, id int, passwordHash *dto.UpdatePasswordRequest) error {
	strQuery := `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2;
	`

	result, err := d.pool.Exec(ctx, strQuery, passwordHash.NewPassword, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}

	return nil
}

func (d *DataBase) UpdateBio(ctx context.Context, id int, bio string) error {
	result, err := d.pool.Exec(ctx, `UPDATE users SET bio = $1 WHERE id = $2`, bio, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (d *DataBase) UpdateSubscriptionStatus(ctx context.Context, id int, sub string) error {
	strQuery := `
		UPDATE users
		SET subscription = $1
		WHERE id = $2;
	`

	result, err := d.pool.Exec(ctx, strQuery, sub, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}

	return nil
}

func (d *DataBase) CreateEmailVerification(ctx context.Context, req *dto.RegisterRequest, token string, expiresAt time.Time) error {
	strQuery := `
		INSERT INTO email_verifications
			(email, nickname, password_hash, birth_date, gender, token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO UPDATE SET
			nickname      = EXCLUDED.nickname,
			password_hash = EXCLUDED.password_hash,
			birth_date    = EXCLUDED.birth_date,
			gender        = EXCLUDED.gender,
			token         = EXCLUDED.token,
			expires_at    = EXCLUDED.expires_at;
	`

	_, err := d.pool.Exec(ctx, strQuery,
		req.Email,
		req.Nickname,
		req.Password,
		req.Birthdate,
		req.Gender,
		token,
		expiresAt,
	)

	return err
}

func (d *DataBase) GetVerificationByToken(ctx context.Context, token string) (*auth.EmailVerification, error) {
	data := &auth.EmailVerification{}

	query := `
		SELECT email, nickname, password_hash, birth_date, gender, expires_at
		FROM email_verifications
		WHERE token = $1
	`

	err := d.pool.QueryRow(ctx, query, token).Scan(
		&data.Email,
		&data.Nickname,
		&data.Password,
		&data.BirthDate,
		&data.Gender,
		&data.Expires,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return data, nil
}

func (d *DataBase) DeleteVerificationByToken(ctx context.Context, token string) error {
	_, err := d.pool.Exec(ctx, `
		DELETE FROM email_verifications
		WHERE token = $1
	`, token)

	return err
}

func (d *DataBase) CreateUser(ctx context.Context, v *auth.EmailVerification) (*dto.RegisterResponse, error) {
	resp := &dto.RegisterResponse{}

	query := `
		INSERT INTO users (
			email, nickname, password_hash, birth_date, gender, is_verified
		)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, email, nickname, birth_date, gender, is_verified
	`

	err := d.pool.QueryRow(ctx, query,
		v.Email,
		v.Nickname,
		v.Password,
		v.BirthDate,
		v.Gender,
	).Scan(
		&resp.ID,
		&resp.Email,
		&resp.Nickname,
		&resp.Birthdate,
		&resp.Gender,
		&resp.IsVerified,
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *DataBase) CreatePasswordReset(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, _ = d.pool.Exec(ctx, `DELETE FROM password_resets WHERE user_id = $1`, userID)

	_, err := d.pool.Exec(ctx, `
        INSERT INTO password_resets (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `, userID, token, expiresAt)

	return err
}

func (d *DataBase) GetPasswordResetByToken(ctx context.Context, token string) (int, time.Time, error) {
	var userID int
	var expiresAt time.Time

	err := d.pool.QueryRow(ctx, `
        SELECT user_id, expires_at
        FROM password_resets
        WHERE token = $1
    `, token).Scan(&userID, &expiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, core.ErrNotFound
		}
		return 0, time.Time{}, err
	}

	return userID, expiresAt, nil
}

func (d *DataBase) DeletePasswordReset(ctx context.Context, token string) error {
	_, err := d.pool.Exec(ctx, `
        DELETE FROM password_resets WHERE token = $1
    `, token)
	return err
}

func (d *DataBase) UpdatePasswordByID(ctx context.Context, userID int, passwordHash string) error {
	result, err := d.pool.Exec(ctx, `
        UPDATE users SET password_hash = $1 WHERE id = $2
    `, passwordHash, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return core.ErrUserNotFound
	}
	return nil
}
