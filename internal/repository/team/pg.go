package team

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zxchelik/avito-test-task/internal/infrastructure/pg"
	"github.com/zxchelik/avito-test-task/internal/model"
	"github.com/zxchelik/avito-test-task/internal/model/team"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// Create inserts a new team by name.
// Returns model.ErrAlreadyExists if team already exists (PK conflict).
func (r *PGRepository) Create(ctx context.Context, name string) error {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
        INSERT INTO teams (name)
        VALUES ($1)
        ON CONFLICT (name) DO NOTHING
    `

	ct, err := q.Exec(ctx, query, name)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		// строка не вставилась => такая команда уже есть
		return model.ErrAlreadyExists
	}

	return nil
}

// GetByName returns a team by team_name.
// Returns model.ErrNotFound if not found.
func (r *PGRepository) GetByName(ctx context.Context, name string) (*team.Team, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `SELECT name FROM teams WHERE name = $1`

	var t team.Team
	err := q.QueryRow(ctx, query, name).Scan(&t.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}
