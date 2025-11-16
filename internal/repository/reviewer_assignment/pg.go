package reviewer_assignment

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zxchelik/avito-test-task/internal/infrastructure/pg"
	"github.com/zxchelik/avito-test-task/internal/model"
	reva "github.com/zxchelik/avito-test-task/internal/model/reviewer_assignment"
	"time"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// ListByPR returns reviewers assigned to a pull request.
func (r *PGRepository) ListByPR(ctx context.Context, prID string) ([]*reva.ReviewerAssignment, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		SELECT pr_id, user_id, assigned_at
		FROM pull_request_reviewers
		WHERE pr_id = $1
		ORDER BY assigned_at
	`

	rows, err := q.Query(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*reva.ReviewerAssignment
	for rows.Next() {
		var ra reva.ReviewerAssignment
		if err := rows.Scan(&ra.PrId, &ra.UserId, &ra.AssignedAt); err != nil {
			return nil, err
		}
		res = append(res, &ra)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Add assigns a reviewer to a PR.
// Returns model.ErrAlreadyExists if conflict occurs (already assigned).
func (r *PGRepository) Add(ctx context.Context, prID, userID string, assignedAt time.Time) error {
	q := pg.GetQuerierFromContext(ctx, r.pool)

	const query = `
		INSERT INTO pull_request_reviewers (pr_id, user_id, assigned_at)
		VALUES ($1, $2, $3)
	`
	ct, err := q.Exec(ctx, query, prID, userID, assignedAt)
	if ct.RowsAffected() == 0 {
		return model.ErrAlreadyExists
	}

	return err
}

// Remove removes reviewer from PR.
// Returns reva.ErrReviewerNotFoundInPR if reviewer not assigned.
func (r *PGRepository) Remove(ctx context.Context, prID, userID string) error {
	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `
		DELETE FROM pull_request_reviewers
		WHERE pr_id = $1 AND user_id = $2
	`

	res, err := q.Exec(ctx, query, prID, userID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return reva.ErrReviewerNotFoundInPR
	}

	return nil
}

// Replace atomically replaces one reviewer with another in a transaction.
// Returns:
//   - reva.ErrReviewerNotFoundInPR — old reviewer wasn't assigned
//   - reva.ErrReviewerSameAsOld — new == old
func (r *PGRepository) Replace(ctx context.Context, prID, oldUserID, newUserID string, assignedAt time.Time) error {
	if oldUserID == newUserID {
		return reva.ErrReviewerSameAsOld
	}

	q := pg.GetQuerierFromContext(ctx, r.pool) // достаём либо tx, либо pool

	const query = `
WITH deleted AS (
    DELETE FROM pull_request_reviewers
    WHERE pr_id = $1 AND user_id = $2
    RETURNING 1
)
INSERT INTO pull_request_reviewers (pr_id, user_id, assigned_at)
SELECT $1, $3, $4
WHERE EXISTS (SELECT 1 FROM deleted)
`

	res, err := q.Exec(ctx, query, prID, oldUserID, newUserID, assignedAt)
	if err != nil {
		return err
	}

	// Если ничего не затронули — значит старого ревьювера не было.
	if res.RowsAffected() == 0 {
		return reva.ErrReviewerNotFoundInPR
	}

	return nil
}

// ListPRIDsByReviewer returns PR IDs where user is assigned as reviewer.
func (r *PGRepository) ListPRIDsByReviewer(ctx context.Context, userID string) ([]string, error) {
	q := pg.GetQuerierFromContext(ctx, r.pool)
	const query = `
		SELECT pr_id
		FROM pull_request_reviewers
		WHERE user_id = $1
		ORDER BY assigned_at DESC
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prIDs, nil
}
