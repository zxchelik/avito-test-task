package user

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zxchelik/avito-test-task/internal/infrastructure/pg"
	"github.com/zxchelik/avito-test-task/internal/model"
	"github.com/zxchelik/avito-test-task/internal/model/user"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// Upsert inserts a new user or updates existing one by ID.
// Returns:
//   - user.ErrTeamNotFound — if related team does not exist (FK violation)
//   - other DB errors
func (r *PGRepository) Upsert(ctx context.Context, u *user.User) error {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		INSERT INTO users (id, name, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    team_name = EXCLUDED.team_name,
		    is_active = EXCLUDED.is_active
		RETURNING id, name, team_name, is_active, created_at
	`

	err := q.QueryRow(ctx, query,
		u.ID, u.Username, u.TeamName, u.IsActive,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt)
	if err != nil {
		// если упали по FK — команды нет
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return user.ErrTeamNotFound
		}
		return err
	}

	return nil
}

// GetByID returns a user by user_id.
// Returns:
//   - model.ErrNotFound — if user not found
func (r *PGRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		SELECT id, name, team_name, is_active, created_at
		FROM users
		WHERE id = $1
	`

	var u user.User
	err := q.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// ListByTeam returns all users belonging to a given team.
func (r *PGRepository) ListByTeam(ctx context.Context, teamName string) ([]*user.User, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		SELECT id, name, team_name, is_active, created_at
		FROM users
		WHERE team_name = $1
		ORDER BY id
	`

	rows, err := q.Query(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User

	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// SetIsActive updates is_active flag for user.
// Returns updated user or model.ErrNotFound if no rows affected.
func (r *PGRepository) SetIsActive(ctx context.Context, id string, isActive bool) (*user.User, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		UPDATE users
		SET is_active = $2
		WHERE id = $1
		RETURNING id, name, team_name, is_active, created_at
	`

	var u user.User
	err := q.QueryRow(ctx, query, id, isActive).Scan(
		&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}
