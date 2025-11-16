package pull_request

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zxchelik/avito-test-task/internal/infrastructure/pg"
	"github.com/zxchelik/avito-test-task/internal/model"
	preq "github.com/zxchelik/avito-test-task/internal/model/pull_request"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// Create inserts a new PR if not exists.
// Returns:
//   - model.ErrAlreadyExists — if PR already exists (PK conflict)
//   - model.ErrNotFound — if author does not exist (FK violation)
func (r *PGRepository) Create(ctx context.Context, pr *preq.PullRequest) (*preq.PullRequest, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `
		INSERT INTO pull_requests (id, title, author_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, title, author_id, status, created_at, merged_at
	`

	err := q.QueryRow(ctx, query,
		pr.ID, pr.Title, pr.AuthorID, pr.Status,
	).Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT с существующим PR
			return nil, model.ErrAlreadyExists
		}

		// FK на author_id
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			// автор не найден
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return pr, nil
}

// GetByID returns a PR by pull_request_id.
// Returns model.ErrNotFound when PR doesn't exist.
func (r *PGRepository) GetByID(ctx context.Context, id string) (*preq.PullRequest, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `
		SELECT id, title, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr preq.PullRequest

	err := q.QueryRow(ctx, query, id).Scan(
		&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

// MarkMerged sets status to MERGED and updates merged_at timestamp.
// Returns:
//   - preq.ErrPRAlreadyMerged — if status is already MERGED
//   - model.ErrNotFound — if PR not found
func (r *PGRepository) MarkMerged(ctx context.Context, id string) (*preq.PullRequest, error) {
	pr, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status == preq.PRMerged {
		return pr, preq.ErrPRAlreadyMerged
	}

	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = NOW()
		WHERE id = $1
		RETURNING id, title, author_id, status, created_at, merged_at
	`

	var newPR preq.PullRequest

	err = q.QueryRow(ctx, query, id).Scan(
		&newPR.ID, &newPR.Title, &newPR.AuthorID, &newPR.Status, &newPR.CreatedAt, &newPR.MergedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// маловероятно (мы только что читали), но формально — not found
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &newPR, nil
}
